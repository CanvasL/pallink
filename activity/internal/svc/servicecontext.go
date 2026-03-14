package svc

import (
	"context"

	"pallink/activity/internal/config"
	"pallink/common/mq"
	"pallink/common/postgres"
	"pallink/user/userclient"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config         config.Config
	DB             *pgxpool.Pool
	UserRpc        userclient.User
	MQ             *mq.Client
	NotificationMQ *mq.Client
}

func NewServiceContext(c config.Config) *ServiceContext {
	pool, err := postgres.NewPool(context.Background(), c.Postgres)
	if err != nil {
		logx.Must(err)
	}
	userCli := zrpc.MustNewClient(c.UserRpc)
	mqClient, err := mq.NewClient(c.AuditMQ)
	if err != nil {
		logx.Must(err)
	}
	notifyClient, err := mq.NewClient(c.NotificationMQ)
	if err != nil {
		logx.Must(err)
	}

	return &ServiceContext{
		Config:         c,
		DB:             pool,
		UserRpc:        userclient.NewUser(userCli),
		MQ:             mqClient,
		NotificationMQ: notifyClient,
	}
}

func (s *ServiceContext) Close() {
	if s.DB != nil {
		s.DB.Close()
	}
	if s.MQ != nil {
		_ = s.MQ.Close()
	}
	if s.NotificationMQ != nil {
		_ = s.NotificationMQ.Close()
	}
}
