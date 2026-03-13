// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"pallink/activity/activityclient"
	"pallink/gateway/internal/config"
	"pallink/notify/notifyclient"
	"pallink/user/userclient"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config      config.Config
	UserRpc     userclient.User
	ActivityRpc activityclient.Activity
	NotifyRpc   notifyclient.Notify
}

func NewServiceContext(c config.Config) *ServiceContext {
	userCli := zrpc.MustNewClient(c.UserRpc)
	activityCli := zrpc.MustNewClient(c.ActivityRpc)
	notifyCli := zrpc.MustNewClient(c.NotifyRpc)
	return &ServiceContext{
		Config:      c,
		UserRpc:     userclient.NewUser(userCli),
		ActivityRpc: activityclient.NewActivity(activityCli),
		NotifyRpc:   notifyclient.NewNotify(notifyCli),
	}
}
