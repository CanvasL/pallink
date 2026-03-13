package svc

import (
	"context"

	"pallink/common/mq"
	"pallink/common/postgres"
	"pallink/notify/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zeromicro/go-zero/core/logx"
)

type ServiceContext struct {
	Config config.Config
	DB     *pgxpool.Pool
	MQ     *mq.Client
}

func NewServiceContext(c config.Config) *ServiceContext {
	pool, err := postgres.NewPool(context.Background(), c.Postgres)
	if err != nil {
		logx.Must(err)
	}
	mqClient, err := mq.NewClient(c.NotifyMQ)
	if err != nil {
		logx.Must(err)
	}

	return &ServiceContext{
		Config: c,
		DB:     pool,
		MQ:     mqClient,
	}
}

func (s *ServiceContext) Close() {
	if s.DB != nil {
		s.DB.Close()
	}
	if s.MQ != nil {
		_ = s.MQ.Close()
	}
}
