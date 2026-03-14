package middleware

import (
	"net"
	"net/http"
	"strconv"
	"strings"

	"pallink/common/auth"
	"pallink/gateway/internal/config"

	"github.com/zeromicro/go-zero/core/limit"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest/httpx"
)

type (
	ruleBucket struct {
		limiter *limit.PeriodLimit
		period  int
	}

	RateLimitMiddleware struct {
		enabled   bool
		failOpen  bool
		login     *ruleBucket
		register  *ruleBucket
		public    *ruleBucket
		userRead  *ruleBucket
		userWrite *ruleBucket
		websocket *ruleBucket
	}
)

func NewRateLimitMiddleware(redisConf redis.RedisConf, cfg config.RateLimitConf) (*RateLimitMiddleware, error) {
	mw := &RateLimitMiddleware{
		enabled:  cfg.Enabled,
		failOpen: cfg.FailOpen,
	}
	if !cfg.Enabled {
		return mw, nil
	}

	store, err := redis.NewRedis(redisConf)
	if err != nil {
		return nil, err
	}

	prefix := cfg.KeyPrefix
	if prefix == "" {
		prefix = "gateway:ratelimit:"
	}

	mw.login = newRuleBucket(cfg.Login, store, prefix+"login:")
	mw.register = newRuleBucket(cfg.Register, store, prefix+"register:")
	mw.public = newRuleBucket(cfg.Public, store, prefix+"public:")
	mw.userRead = newRuleBucket(cfg.UserRead, store, prefix+"user-read:")
	mw.userWrite = newRuleBucket(cfg.UserWrite, store, prefix+"user-write:")
	mw.websocket = newRuleBucket(cfg.Websocket, store, prefix+"websocket:")

	return mw, nil
}

func (m *RateLimitMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	if m == nil || !m.enabled {
		return next
	}

	return func(w http.ResponseWriter, r *http.Request) {
		bucket, key := m.resolveBucket(r)
		if bucket == nil {
			next(w, r)
			return
		}

		code, err := bucket.limiter.TakeCtx(r.Context(), key)
		if err != nil {
			logc.Errorf(r.Context(), "rate limit check failed for %s %s: %v", r.Method, r.URL.Path, err)
			if m.failOpen {
				next(w, r)
				return
			}

			httpx.WriteJsonCtx(r.Context(), w, http.StatusServiceUnavailable, map[string]any{
				"success": false,
				"message": "rate limiter unavailable",
			})
			return
		}

		switch code {
		case limit.Allowed, limit.HitQuota:
			next(w, r)
		case limit.OverQuota:
			if bucket.period > 0 {
				w.Header().Set("Retry-After", strconv.Itoa(bucket.period))
			}
			httpx.WriteJsonCtx(r.Context(), w, http.StatusTooManyRequests, map[string]any{
				"success": false,
				"message": "too many requests",
			})
		default:
			logc.Errorf(r.Context(), "rate limiter returned unknown code for %s %s", r.Method, r.URL.Path)
			if m.failOpen {
				next(w, r)
				return
			}

			httpx.WriteJsonCtx(r.Context(), w, http.StatusServiceUnavailable, map[string]any{
				"success": false,
				"message": "rate limiter unavailable",
			})
		}
	}
}

func newRuleBucket(rule config.RateLimitRule, store *redis.Redis, prefix string) *ruleBucket {
	if store == nil || rule.Period <= 0 || rule.Quota <= 0 {
		return nil
	}

	return &ruleBucket{
		limiter: limit.NewPeriodLimit(rule.Period, rule.Quota, store, prefix),
		period:  rule.Period,
	}
}

func (m *RateLimitMiddleware) resolveBucket(r *http.Request) (*ruleBucket, string) {
	if r == nil {
		return nil, ""
	}

	if r.Method == http.MethodOptions {
		return nil, ""
	}

	path := normalizePath(r.URL.Path)
	if path == "" || path == "/metrics" {
		return nil, ""
	}

	routeKey := routeKey(r.Method, path)
	clientIP := requestIP(r)

	switch {
	case path == "/user/login":
		return m.login, clientIP + ":" + routeKey
	case path == "/user/register":
		return m.register, clientIP + ":" + routeKey
	case path == "/im/ws":
		return m.websocket, clientIP + ":" + routeKey
	case strings.HasPrefix(path, "/activity/public/"):
		return m.public, clientIP + ":" + routeKey
	}

	userID, ok := auth.GetUserIDFromCtx(r.Context())
	identity := clientIP
	if ok && userID > 0 {
		identity = "u" + strconv.FormatUint(userID, 10)
	}

	switch r.Method {
	case http.MethodGet, http.MethodHead:
		return m.userRead, identity + ":" + routeKey
	default:
		return m.userWrite, identity + ":" + routeKey
	}
}

func normalizePath(path string) string {
	if path == "" {
		return ""
	}

	if path != "/" {
		path = strings.TrimRight(path, "/")
	}

	return path
}

func routeKey(method, path string) string {
	replacer := strings.NewReplacer("/", ":", " ", "", "\t", "", "\n", "")
	return strings.ToLower(method) + ":" + replacer.Replace(path)
}

func requestIP(r *http.Request) string {
	for _, header := range []string{"X-Forwarded-For", "X-Real-IP"} {
		value := strings.TrimSpace(r.Header.Get(header))
		if value == "" {
			continue
		}

		if header == "X-Forwarded-For" {
			parts := strings.Split(value, ",")
			value = strings.TrimSpace(parts[0])
		}

		if value != "" {
			return value
		}
	}

	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil && host != "" {
		return host
	}

	return strings.TrimSpace(r.RemoteAddr)
}
