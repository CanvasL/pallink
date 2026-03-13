package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"pallink/activity/activityclient"
	"pallink/audit/internal/config"
	"pallink/common/mq"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
)

var configFile = flag.String("f", "etc/audit.yaml", "the config file")

type auditMessage struct {
	ActivityID uint64 `json:"activity_id"`
}

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	mqClient, err := mq.NewClient(c.RabbitMQ)
	if err != nil {
		logx.Must(err)
	}
	defer mqClient.Close()

	activityCli := zrpc.MustNewClient(c.ActivityRpc)
	activitySvc := activityclient.NewActivity(activityCli)

	msgs, err := mqClient.Consume()
	if err != nil {
		logx.Must(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logx.Infof("audit worker started, queue=%s", c.RabbitMQ.Queue)

	for {
		select {
		case <-ctx.Done():
			logx.Info("audit worker stopping")
			return
		case msg, ok := <-msgs:
			if !ok {
				logx.Error("rabbitmq channel closed")
				return
			}

			var body auditMessage
			if err := json.Unmarshal(msg.Body, &body); err != nil {
				_ = msg.Nack(false, false)
				continue
			}
			if body.ActivityID == 0 {
				_ = msg.Nack(false, false)
				continue
			}

			_, err := activitySvc.UpdateAuditStatus(ctx, &activityclient.UpdateAuditStatusRequest{
				ActivityId:  body.ActivityID,
				AuditStatus: 1,
			})
			if err != nil {
				_ = msg.Nack(false, true)
				continue
			}

			_ = msg.Ack(false)
			fmt.Printf("audited activity=%d -> approved\n", body.ActivityID)
		}
	}
}
