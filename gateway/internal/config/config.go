// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package config

import (
	"pallink/common/mq"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf
	Auth struct {
		AccessSecret string
		AccessExpire int64
	}
	UserRpc         zrpc.RpcClientConf
	ActivityRpc     zrpc.RpcClientConf
	NotificationRpc zrpc.RpcClientConf
	ImRpc           zrpc.RpcClientConf
	RealtimeMQ      mq.FanoutConfig
	Redis           RedisConf
	RateLimit       RateLimitConf
}

type RedisConf struct {
	RateLimit       redis.RedisConf
	Cache           redis.RedisConf
	DistributedLock redis.RedisConf
}

type RateLimitConf struct {
	Enabled   bool
	FailOpen  bool
	KeyPrefix string
	Login     RateLimitRule
	Register  RateLimitRule
	Public    RateLimitRule
	UserRead  RateLimitRule
	UserWrite RateLimitRule
	Websocket RateLimitRule
}

type RateLimitRule struct {
	Period int
	Quota  int
}
