package config

import (
	"pallink/common/mq"

	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	Name        string
	AuditMQ     mq.Config
	ActivityRpc zrpc.RpcClientConf
	UserRpc     zrpc.RpcClientConf
}
