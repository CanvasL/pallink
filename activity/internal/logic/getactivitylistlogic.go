package logic

import (
	"context"

	"pallink/activity/activity"
	"pallink/activity/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetActivityListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetActivityListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetActivityListLogic {
	return &GetActivityListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetActivityListLogic) GetActivityList(in *activity.GetActivityListRequest) (*activity.GetActivityListResponse, error) {
	filter := listFilter{
		Status:  nil,
		Keyword: in.Keyword,
	}
	if in.Status != 0 {
		filter.Status = &in.Status
	}
	if in.ViewerUserId == 0 {
		approved := int32(1)
		filter.AuditStatus = &approved
	}

	activities, total, err := queryActivityList(l.ctx, l.svcCtx.DB, filter, in.ViewerUserId, in.Page, in.PageSize)
	if err != nil {
		return nil, err
	}
	if err := hydrateActivityUsers(l.ctx, l.svcCtx.UserRpc, activities); err != nil {
		return nil, err
	}

	return &activity.GetActivityListResponse{
		Activities: activities,
		Total:      total,
	}, nil
}
