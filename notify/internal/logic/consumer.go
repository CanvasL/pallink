package logic

import (
	"context"
	"encoding/json"
	"strings"

	"pallink/common/mq"
	"pallink/notify/internal/dao"
	"pallink/notify/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

func StartConsumer(ctx context.Context, svcCtx *svc.ServiceContext) {
	msgs, err := svcCtx.MQ.Consume()
	if err != nil {
		logx.Error(err)
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-msgs:
			if !ok {
				logx.Error("rabbitmq channel closed")
				return
			}

			var evt mq.CommentNotifyEvent
			if err := json.Unmarshal(msg.Body, &evt); err != nil {
				_ = msg.Nack(false, false)
				continue
			}
			if evt.CommentId == 0 || evt.ActivityId == 0 || evt.ActorId == 0 {
				_ = msg.Nack(false, false)
				continue
			}

			content := strings.TrimSpace(evt.Content)
			if len(content) > 80 {
				content = content[:80]
			}

			// recipients: activity creator + parent comment user
			recipients := make(map[uint64]struct{})
			if evt.ActivityCreatorId > 0 && evt.ActivityCreatorId != evt.ActorId {
				recipients[evt.ActivityCreatorId] = struct{}{}
			}
			if evt.ParentUserId > 0 && evt.ParentUserId != evt.ActorId {
				recipients[evt.ParentUserId] = struct{}{}
			}

			for uid := range recipients {
				typ := "activity_comment"
				if evt.ParentId > 0 {
					typ = "comment_reply"
				}
				err := dao.InsertNotification(ctx, svcCtx.DB, uid, evt.ActorId, typ, evt.ActivityId, evt.CommentId, evt.ParentId, content)
				if err != nil {
					_ = msg.Nack(false, true)
					goto done
				}
			}

			_ = msg.Ack(false)
		done:
		}
	}
}
