package logic

import (
	"context"
	"errors"

	"pallink/activity/rpc/activity"
	"pallink/activity/rpc/internal/svc"

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

	var (
		id uint64
	)
	err := l.svcCtx.DB.QueryRow(
		l.ctx,
		`INSERT INTO activity (creator_id, title, description, location, start_time, end_time, max_people, current_people, status, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,0,1,now())
		 RETURNING id`,
		in.CreatorId, in.Title, in.Description, in.Location, startTime, endTime, in.MaxPeople,
	).Scan(&id)
	if err != nil {
		return nil, err
	}

	return queryActivityDetail(l.ctx, l.svcCtx.DB, id, in.CreatorId)
}
