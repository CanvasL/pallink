package logic

import (
	"context"
	"errors"

	"pallink/im/im"
	"pallink/im/internal/dao"
	"pallink/im/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetConversationListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetConversationListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetConversationListLogic {
	return &GetConversationListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetConversationListLogic) GetConversationList(in *im.GetConversationListRequest) (*im.GetConversationListResponse, error) {
	if in.UserId == 0 {
		return nil, errors.New("user_id required")
	}

	list, total, err := dao.QueryConversationList(l.ctx, l.svcCtx.DB, in.UserId, in.Page, in.PageSize)
	if err != nil {
		return nil, err
	}
	if err := hydrateConversations(l.ctx, l.svcCtx.UserRpc, list); err != nil {
		return nil, err
	}

	return &im.GetConversationListResponse{
		List:  list,
		Total: total,
	}, nil
}
