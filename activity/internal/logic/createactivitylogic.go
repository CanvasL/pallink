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

type CreateActivityLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateActivityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateActivityLogic {
	return &CreateActivityLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateActivityLogic) CreateActivity(in *activity.CreateActivityRequest) (*activity.ActivityInfo, error) {
	if in.CreatorId == 0 {
		return nil, errors.New("creator_id required")
	}
	if in.Title == "" || in.Location == "" {
		return nil, errors.New("title/location required")
	}
	if in.StartTime == nil || in.EndTime == nil {
		return nil, errors.New("start_time/end_time required")
	}
	startTime := in.StartTime.AsTime()
	endTime := in.EndTime.AsTime()
	if !endTime.After(startTime) {
		return nil, errors.New("end_time must be after start_time")
	}
	if in.MaxPeople < 0 {
		return nil, errors.New("max_people invalid")
	}

	id, err := dao.CreateActivity(l.ctx, l.svcCtx.DB, in.CreatorId, in.Title, in.Description, in.Location, startTime, endTime, in.MaxPeople)
	if err != nil {
		return nil, err
	}
	if err := l.svcCtx.MQ.PublishJSON(l.ctx, mq.AuditMessage{Type: "activity", ID: id}); err != nil {
		return nil, err
	}

	info, err := dao.QueryActivityDetail(l.ctx, l.svcCtx.DB, id, in.CreatorId)
	if err != nil {
		return nil, err
	}
	if err := hydrateActivityUsers(l.ctx, l.svcCtx.UserRpc, []*activity.ActivityInfo{info}); err != nil {
		return nil, err
	}
	return info, nil
}
