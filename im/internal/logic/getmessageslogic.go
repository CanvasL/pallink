package logic

import (
	"context"
	"errors"

	"pallink/im/im"
	"pallink/im/internal/dao"
	"pallink/im/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMessagesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMessagesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMessagesLogic {
	return &GetMessagesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetMessagesLogic) GetMessages(in *im.GetMessagesRequest) (*im.GetMessagesResponse, error) {
	if in.UserId == 0 || in.ConversationId == 0 {
		return nil, errors.New("user_id/conversation_id required")
	}

	list, hasMore, err := dao.QueryMessages(l.ctx, l.svcCtx.DB, in.UserId, in.ConversationId, in.BeforeMessageId, in.PageSize)
	if err != nil {
		return nil, err
	}
	return &im.GetMessagesResponse{
		List:    list,
		HasMore: hasMore,
	}, nil
}
