package logic

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"pallink/common/mq"
	"pallink/notification/internal/dao"
	"pallink/notification/internal/svc"

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

			if err := handleNotificationMessage(ctx, svcCtx, msg.Body); err != nil {
				_ = msg.Nack(false, true)
				continue
			}
			_ = msg.Ack(false)
		}
	}
}

func handleNotificationMessage(ctx context.Context, svcCtx *svc.ServiceContext, body []byte) error {
	var probe struct {
		CommentId uint64 `json:"comment_id"`
		MessageId uint64 `json:"message_id"`
	}
	if err := json.Unmarshal(body, &probe); err != nil {
		return err
	}

	switch {
	case probe.CommentId > 0:
		return handleCommentNotification(ctx, svcCtx, body)
	case probe.MessageId > 0:
		return handleImMessageNotification(ctx, svcCtx, body)
	default:
		return errors.New("unknown notification message")
	}
}

func handleCommentNotification(ctx context.Context, svcCtx *svc.ServiceContext, body []byte) error {
	var evt mq.CommentNotificationEvent
	if err := json.Unmarshal(body, &evt); err != nil {
		return err
	}
	if evt.CommentId == 0 || evt.ActivityId == 0 || evt.ActorId == 0 {
		return errors.New("invalid comment notification event")
	}

	content := trimContent(evt.Content)

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
		if err := dao.InsertNotification(ctx, svcCtx.DB, uid, evt.ActorId, typ, evt.ActivityId, evt.CommentId, evt.ParentId, content); err != nil {
			return err
		}
	}
	return nil
}

func handleImMessageNotification(ctx context.Context, svcCtx *svc.ServiceContext, body []byte) error {
	var evt mq.ImMessageNotificationEvent
	if err := json.Unmarshal(body, &evt); err != nil {
		return err
	}
	if evt.MessageId == 0 || evt.ActorId == 0 || evt.ReceiverId == 0 {
		return errors.New("invalid im notification event")
	}

	return dao.InsertNotification(ctx, svcCtx.DB, evt.ReceiverId, evt.ActorId, "im_message", 0, 0, 0, trimContent(evt.Content))
}

func trimContent(content string) string {
	content = strings.TrimSpace(content)
	if len(content) > 80 {
		content = content[:80]
	}
	return content
}
