package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
)

// GetUserIDFromCtx extracts userId from context (set by go-zero JWT middleware).
func GetUserIDFromCtx(ctx context.Context) (uint64, bool) {
	if ctx == nil {
		return 0, false
	}

	val := ctx.Value("userId")
	switch v := val.(type) {
	case uint64:
		return v, true
	case int64:
		if v < 0 {
			return 0, false
		}
		return uint64(v), true
	case int:
		if v < 0 {
			return 0, false
		}
		return uint64(v), true
	case float64:
		if v < 0 {
			return 0, false
		}
		return uint64(v), true
	case json.Number:
		u, err := strconv.ParseUint(v.String(), 10, 64)
		if err != nil {
			return 0, false
		}
		return u, true
	case string:
		u, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return 0, false
		}
		return u, true
	case fmt.Stringer:
		u, err := strconv.ParseUint(v.String(), 10, 64)
		if err != nil {
			return 0, false
		}
		return u, true
	default:
		return 0, false
	}
}
