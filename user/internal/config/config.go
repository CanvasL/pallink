package config

import (
	"pallink/common/mq"
	"pallink/common/postgres"

	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	Postgres postgres.Config
	Jwt      struct {
		AccessSecret string
		AccessExpire int64
	}
	RabbitMQ mq.Config
}
