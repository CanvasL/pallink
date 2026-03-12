// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"pallink/user/api/internal/config"
	"pallink/user/rpc/userclient"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config  config.Config
	UserRpc userclient.User
}

func NewServiceContext(c config.Config) *ServiceContext {
	client := zrpc.MustNewClient(c.UserRpc)
	return &ServiceContext{
		Config:  c,
		UserRpc: userclient.NewUser(client),
	}
}
