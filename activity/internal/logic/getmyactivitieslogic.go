package logic

import (
	"context"
	"errors"

	"pallink/activity/activity"
	"pallink/activity/internal/dao"
	"pallink/activity/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMyActivitiesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMyActivitiesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMyActivitiesLogic {
	return &GetMyActivitiesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetMyActivitiesLogic) GetMyActivities(in *activity.GetMyActivitiesRequest) (*activity.GetActivityListResponse, error) {
	if in.UserId == 0 {
		return nil, errors.New("user_id required")
	}

	filter := dao.ListFilter{
		CreatorID:    in.UserId,
		UseCreatorID: true,
	}

	activities, total, err := dao.QueryActivityList(l.ctx, l.svcCtx.DB, filter, in.UserId, in.Page, in.PageSize)
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
