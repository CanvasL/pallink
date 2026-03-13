package logic

import (
	"context"
	"errors"

	"pallink/activity/activity"
	"pallink/activity/internal/dao"
	"pallink/activity/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetEnrolledActivitiesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetEnrolledActivitiesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetEnrolledActivitiesLogic {
	return &GetEnrolledActivitiesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetEnrolledActivitiesLogic) GetEnrolledActivities(in *activity.GetEnrolledActivitiesRequest) (*activity.GetActivityListResponse, error) {
	if in.UserId == 0 {
		return nil, errors.New("user_id required")
	}

	filter := dao.ListFilter{
		EnrolledUserID:  in.UserId,
		UseEnrolledUser: true,
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
