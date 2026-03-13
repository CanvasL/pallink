package config

import (
	"pallink/common/mq"
	"pallink/common/postgres"

	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	Postgres postgres.Config
	UserRpc  zrpc.RpcClientConf
	AuditMQ  mq.Config
	NotifyMQ mq.Config
}
