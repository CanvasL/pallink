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
	"pallink/user/userclient"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
)

var configFile = flag.String("f", "etc/audit.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	mqClient, err := mq.NewClient(c.AuditMQ)
	if err != nil {
		logx.Must(err)
	}
	defer mqClient.Close()

	activityCli := zrpc.MustNewClient(c.ActivityRpc)
	activitySvc := activityclient.NewActivity(activityCli)
	userCli := zrpc.MustNewClient(c.UserRpc)
	userSvc := userclient.NewUser(userCli)

	msgs, err := mqClient.Consume()
	if err != nil {
		logx.Must(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logx.Infof("audit worker started, queue=%s", c.AuditMQ.Queue)

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

			var body mq.AuditMessage
			if err := json.Unmarshal(msg.Body, &body); err != nil {
				_ = msg.Nack(false, false)
				continue
			}
			if body.ID == 0 || body.Type == "" {
				_ = msg.Nack(false, false)
				continue
			}

			switch body.Type {
			case "activity":
				_, err := activitySvc.UpdateAuditStatus(ctx, &activityclient.UpdateAuditStatusRequest{
					ActivityId:  body.ID,
					AuditStatus: 1,
				})
				if err != nil {
					_ = msg.Nack(false, true)
					continue
				}
			case "comment":
				_, err := activitySvc.UpdateCommentAuditStatus(ctx, &activityclient.UpdateCommentAuditStatusRequest{
					CommentId:   body.ID,
					AuditStatus: 1,
				})
				if err != nil {
					_ = msg.Nack(false, true)
					continue
				}
			case "user":
				_, err := userSvc.UpdateUserAuditStatus(ctx, &userclient.UpdateUserAuditStatusRequest{
					UserId:      body.ID,
					AuditStatus: 1,
				})
				if err != nil {
					_ = msg.Nack(false, true)
					continue
				}
			default:
				_ = msg.Nack(false, false)
				continue
			}

			_ = msg.Ack(false)
			fmt.Printf("audited %s=%d -> approved\n", body.Type, body.ID)
		}
	}
}
