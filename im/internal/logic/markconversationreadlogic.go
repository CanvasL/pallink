package logic

import (
	"context"
	"errors"

	"pallink/im/im"
	"pallink/im/internal/dao"
	"pallink/im/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type MarkConversationReadLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMarkConversationReadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MarkConversationReadLogic {
	return &MarkConversationReadLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *MarkConversationReadLogic) MarkConversationRead(in *im.MarkConversationReadRequest) (*im.MarkConversationReadResponse, error) {
	if in.UserId == 0 || in.ConversationId == 0 {
		return nil, errors.New("user_id/conversation_id required")
	}

	lastReadMsgID, err := dao.MarkConversationRead(l.ctx, l.svcCtx.DB, in.UserId, in.ConversationId, in.LastReadMsgId)
	if err != nil {
		return nil, err
	}
	return &im.MarkConversationReadResponse{
		Success:       true,
		Message:       "ok",
		LastReadMsgId: lastReadMsgID,
	}, nil
}
