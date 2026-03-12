package svc

import (
	"context"

	"pallink/common/postgres"
	"pallink/user/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zeromicro/go-zero/core/logx"
)

type ServiceContext struct {
	Config config.Config
	DB     *pgxpool.Pool
}

func NewServiceContext(c config.Config) *ServiceContext {
	pool, err := postgres.NewPool(context.Background(), c.Postgres)
	if err != nil {
		logx.Must(err)
	}

	return &ServiceContext{
		Config: c,
		DB:     pool,
	}
}

func (s *ServiceContext) Close() {
	if s.DB != nil {
		s.DB.Close()
	}
}
