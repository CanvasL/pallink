// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"pallink/activity/api/internal/config"
	"pallink/activity/rpc/activityclient"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config      config.Config
	ActivityRpc activityclient.Activity
}

func NewServiceContext(c config.Config) *ServiceContext {
	client := zrpc.MustNewClient(c.ActivityRpc)
	return &ServiceContext{
		Config:      c,
		ActivityRpc: activityclient.NewActivity(client),
	}
}
