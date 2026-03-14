package svc

import (
	"context"

	"pallink/common/mq"
	"pallink/common/postgres"
	"pallink/im/internal/config"
	"pallink/user/userclient"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config     config.Config
	DB         *pgxpool.Pool
	UserRpc    userclient.User
	NotifyMQ   *mq.Client
	RealtimeMQ *mq.FanoutPublisher
}

func NewServiceContext(c config.Config) *ServiceContext {
	pool, err := postgres.NewPool(context.Background(), c.Postgres)
	if err != nil {
		logx.Must(err)
	}
	userCli := zrpc.MustNewClient(c.UserRpc)
	notifyMQ, err := mq.NewClient(c.NotifyMQ)
	if err != nil {
		logx.Must(err)
	}
	realtimeMQ, err := mq.NewFanoutPublisher(c.RealtimeMQ)
	if err != nil {
		logx.Must(err)
	}

	return &ServiceContext{
		Config:     c,
		DB:         pool,
		UserRpc:    userclient.NewUser(userCli),
		NotifyMQ:   notifyMQ,
		RealtimeMQ: realtimeMQ,
	}
}

func (s *ServiceContext) Close() {
	if s.DB != nil {
		s.DB.Close()
	}
	if s.NotifyMQ != nil {
		_ = s.NotifyMQ.Close()
	}
	if s.RealtimeMQ != nil {
		_ = s.RealtimeMQ.Close()
	}
}
