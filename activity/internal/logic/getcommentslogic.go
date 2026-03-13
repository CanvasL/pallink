package logic

import (
	"context"
	"errors"

	"pallink/activity/activity"
	"pallink/activity/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCommentsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCommentsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCommentsLogic {
	return &GetCommentsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetCommentsLogic) GetComments(in *activity.GetCommentsRequest) (*activity.GetCommentsResponse, error) {
	if in.ActivityId == 0 {
		return nil, errors.New("activity_id required")
	}
	if in.ParentId > 0 && in.ActivityId == 0 {
		return nil, errors.New("activity_id required")
	}

	list, total, err := queryComments(l.ctx, l.svcCtx.DB, in.ActivityId, in.ParentId, in.ViewerUserId, in.Page, in.PageSize)
	if err != nil {
		return nil, err
	}
	if err := hydrateComments(l.ctx, l.svcCtx.UserRpc, list); err != nil {
		return nil, err
	}

	return &activity.GetCommentsResponse{Comments: list, Total: total}, nil
}
