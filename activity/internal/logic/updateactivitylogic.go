package logic

import (
	"context"
	"errors"

	"pallink/activity/activity"
	"pallink/activity/internal/dao"
	"pallink/activity/internal/svc"
	"pallink/common/mq"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateActivityLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateActivityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateActivityLogic {
	return &UpdateActivityLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateActivityLogic) UpdateActivity(in *activity.UpdateActivityRequest) (*activity.ActivityInfo, error) {
	if in.Id == 0 || in.CreatorId == 0 {
		return nil, errors.New("id/creator_id required")
	}

	if in.Title == "" && in.Description == "" && in.Location == "" && in.StartTime == nil && in.EndTime == nil && in.MaxPeople == -1 && in.Status == -1 {
		return nil, errors.New("no fields to update")
	}

	if in.StartTime != nil && in.EndTime != nil {
		if !in.EndTime.AsTime().After(in.StartTime.AsTime()) {
			return nil, errors.New("end_time must be after start_time")
		}
	}

	updated, err := dao.UpdateActivity(l.ctx, l.svcCtx.DB, in)
	if err != nil {
		return nil, err
	}
	if !updated {
		return nil, errors.New("activity not found or forbidden")
	}

	if err := l.svcCtx.MQ.PublishJSON(l.ctx, mq.AuditMessage{Type: "activity", ID: in.Id}); err != nil {
		return nil, err
	}

	info, err := dao.QueryActivityDetail(l.ctx, l.svcCtx.DB, in.Id, in.CreatorId)
	if err != nil {
		return nil, err
	}
	if err := hydrateActivityUsers(l.ctx, l.svcCtx.UserRpc, []*activity.ActivityInfo{info}); err != nil {
		return nil, err
	}
	return info, nil
}
