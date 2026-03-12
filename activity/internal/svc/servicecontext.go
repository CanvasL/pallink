package svc

import (
	"context"

	"pallink/activity/internal/config"
	"pallink/common/postgres"
	"pallink/user/userclient"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config  config.Config
	DB      *pgxpool.Pool
	UserRpc userclient.User
}

func NewServiceContext(c config.Config) *ServiceContext {
	pool, err := postgres.NewPool(context.Background(), c.Postgres)
	if err != nil {
		logx.Must(err)
	}
	userCli := zrpc.MustNewClient(c.UserRpc)

	return &ServiceContext{
		Config:  c,
		DB:      pool,
		UserRpc: userclient.NewUser(userCli),
	}
}

func (s *ServiceContext) Close() {
	if s.DB != nil {
		s.DB.Close()
	}
}
