// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package main

import (
	"context"
	"flag"
	"fmt"

	"pallink/gateway/internal/config"
	"pallink/gateway/internal/handler"
	"pallink/gateway/internal/svc"
	gatewayws "pallink/gateway/internal/ws"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/gatewayapi.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(
		c.RestConf,
		rest.WithCors(),
		rest.WithCorsHeaders("Authorization", "Content-Type", "X-Requested-With"),
	)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	defer ctx.Close()
	if ctx.RateLimiter != nil {
		server.Use(ctx.RateLimiter.Handle)
	}
	runCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go gatewayws.StartRealtimeConsumer(runCtx, ctx.RealtimeMQ, ctx.WsHub)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
