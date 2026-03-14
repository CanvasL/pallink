// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package config

import (
	"pallink/common/mq"

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
}
