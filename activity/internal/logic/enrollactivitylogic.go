package logic

import (
	"context"
	"errors"
	"time"

	"pallink/activity/activity"
	"pallink/activity/internal/dao"
	"pallink/activity/internal/svc"

	"github.com/jackc/pgx/v5"
	"github.com/zeromicro/go-zero/core/logx"
)

type EnrollActivityLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewEnrollActivityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *EnrollActivityLogic {
	return &EnrollActivityLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *EnrollActivityLogic) EnrollActivity(in *activity.EnrollActivityRequest) (*activity.EnrollActivityResponse, error) {
	if in.ActivityId == 0 || in.UserId == 0 {
		return &activity.EnrollActivityResponse{Success: false, Message: "activity_id/user_id required"}, nil
	}

	tx, err := l.svcCtx.DB.Begin(l.ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(l.ctx)

	maxPeople, currentPeople, status, err := dao.GetActivityEnrollInfoForUpdate(l.ctx, tx, in.ActivityId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &activity.EnrollActivityResponse{Success: false, Message: "activity not found"}, nil
		}
		return nil, err
	}
	if status != 1 {
		return &activity.EnrollActivityResponse{Success: false, Message: "activity not open for enrollment"}, nil
	}
	if maxPeople > 0 && currentPeople >= maxPeople {
		return &activity.EnrollActivityResponse{Success: false, Message: "activity is full"}, nil
	}

	existingStatus, exists, err := dao.GetEnrollmentStatusForUpdate(l.ctx, tx, in.ActivityId, in.UserId)
	if err != nil {
		return nil, err
	}
	if exists {
		if existingStatus == 1 || existingStatus == 2 {
			return &activity.EnrollActivityResponse{Success: false, Message: "already enrolled"}, nil
		}
		now := time.Now()
		if err := dao.UpdateEnrollmentStatus(l.ctx, tx, in.ActivityId, in.UserId, 1, &now, nil); err != nil {
			return nil, err
		}
	} else {
		if err := dao.InsertEnrollment(l.ctx, tx, in.ActivityId, in.UserId, 1, time.Now()); err != nil {
			return nil, err
		}
	}

	if err := dao.UpdateActivityPeople(l.ctx, tx, in.ActivityId, 1); err != nil {
		return nil, err
	}

	if err := tx.Commit(l.ctx); err != nil {
		return nil, err
	}

	return &activity.EnrollActivityResponse{Success: true, Message: "ok"}, nil
}
