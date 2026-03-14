package logic

import (
	"context"
	"errors"
	"strings"

	"pallink/common/mq"
	"pallink/im/im"
	"pallink/im/internal/dao"
	"pallink/im/internal/svc"
	"pallink/user/userclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type SendMessageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSendMessageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendMessageLogic {
	return &SendMessageLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SendMessageLogic) SendMessage(in *im.SendMessageRequest) (*im.MessageInfo, error) {
	if in.SenderId == 0 || in.PeerUserId == 0 {
		return nil, errors.New("sender_id/peer_user_id required")
	}
	if in.SenderId == in.PeerUserId {
		return nil, errors.New("cannot chat with self")
	}
	content := strings.TrimSpace(in.Content)
	if content == "" {
		return nil, errors.New("content required")
	}

	if _, err := l.svcCtx.UserRpc.GetUserInfo(l.ctx, &userclient.GetUserInfoRequest{UserId: in.PeerUserId}); err != nil {
		return nil, err
	}

	tx, err := l.svcCtx.DB.Begin(l.ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(l.ctx)

	msg, err := dao.CreateMessage(l.ctx, tx, in.SenderId, in.PeerUserId, content)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(l.ctx); err != nil {
		return nil, err
	}

	if err := l.svcCtx.NotificationMQ.PublishJSON(l.ctx, mq.ImMessageNotificationEvent{
		MessageId:      msg.Id,
		ConversationId: msg.ConversationId,
		ActorId:        in.SenderId,
		ReceiverId:     in.PeerUserId,
		Content:        content,
	}); err != nil {
		return nil, err
	}
	if err := l.svcCtx.RealtimeMQ.PublishJSON(l.ctx, mq.ImRealtimeEvent{
		Type:           "im_message",
		Targets:        []uint64{in.SenderId, in.PeerUserId},
		MessageId:      msg.Id,
		ConversationId: msg.ConversationId,
		SenderId:       in.SenderId,
		ReceiverId:     in.PeerUserId,
		Content:        msg.Content,
		AuditStatus:    msg.AuditStatus,
		CreatedAt:      msg.CreatedAt.AsTime().Unix(),
	}); err != nil {
		l.Errorf("publish realtime im event failed: %v", err)
	}
	return msg, nil
}
