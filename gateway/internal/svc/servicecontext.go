// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"pallink/activity/activityclient"
	"pallink/common/mq"
	"pallink/gateway/internal/config"
	"pallink/gateway/internal/middleware"
	gatewayws "pallink/gateway/internal/ws"
	"pallink/im/imclient"
	"pallink/notification/notificationclient"
	"pallink/user/userclient"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config          config.Config
	UserRpc         userclient.User
	ActivityRpc     activityclient.Activity
	NotificationRpc notificationclient.Notification
	ImRpc           imclient.Im
	RealtimeMQ      *mq.FanoutSubscriber
	RateLimiter     *middleware.RateLimitMiddleware
	WsHub           *gatewayws.Hub
}

func NewServiceContext(c config.Config) *ServiceContext {
	userCli := zrpc.MustNewClient(c.UserRpc)
	activityCli := zrpc.MustNewClient(c.ActivityRpc)
	notificationCli := zrpc.MustNewClient(c.NotificationRpc)
	imCli := zrpc.MustNewClient(c.ImRpc)
	rateLimiter, err := middleware.NewRateLimitMiddleware(c.Redis.RateLimit, c.RateLimit)
	if err != nil {
		logx.Must(err)
	}
	realtimeMQ, err := mq.NewFanoutSubscriber(c.RealtimeMQ)
	if err != nil {
		logx.Must(err)
	}
	return &ServiceContext{
		Config:          c,
		UserRpc:         userclient.NewUser(userCli),
		ActivityRpc:     activityclient.NewActivity(activityCli),
		NotificationRpc: notificationclient.NewNotification(notificationCli),
		ImRpc:           imclient.NewIm(imCli),
		RealtimeMQ:      realtimeMQ,
		RateLimiter:     rateLimiter,
		WsHub:           gatewayws.NewHub(),
	}
}

func (s *ServiceContext) Close() {
	if s.RealtimeMQ != nil {
		_ = s.RealtimeMQ.Close()
	}
	if s.WsHub != nil {
		s.WsHub.Close()
	}
}
