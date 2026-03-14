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

type CheckInLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCheckInLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckInLogic {
	return &CheckInLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CheckInLogic) CheckIn(in *activity.CheckInRequest) (*activity.EnrollActivityResponse, error) {
	if in.ActivityId == 0 || in.UserId == 0 {
		return &activity.EnrollActivityResponse{Success: false, Message: "activity_id/user_id required"}, nil
	}

	startTime, activityStatus, err := dao.GetActivityCheckInInfo(l.ctx, l.svcCtx.DB, in.ActivityId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &activity.EnrollActivityResponse{Success: false, Message: "activity not found"}, nil
		}
		return nil, err
	}

	status, exists, err := dao.GetEnrollmentStatus(l.ctx, l.svcCtx.DB, in.ActivityId, in.UserId)
	if err != nil {
		return nil, err
	}
	if message := validateCheckIn(time.Now(), startTime, activityStatus, status, exists); message != "" {
		return &activity.EnrollActivityResponse{Success: false, Message: message}, nil
	}

	now := time.Now()
	if err := dao.UpdateEnrollmentStatus(l.ctx, l.svcCtx.DB, in.ActivityId, in.UserId, 2, nil, &now); err != nil {
		return nil, err
	}

	return &activity.EnrollActivityResponse{Success: true, Message: "ok"}, nil
}

func validateCheckIn(now, startTime time.Time, activityStatus, enrollStatus int32, enrolled bool) string {
	if activityStatus != 1 {
		return "activity not open for checkin"
	}
	if now.Before(startTime) {
		return "activity has not started"
	}
	if !enrolled {
		return "not enrolled"
	}
	if enrollStatus == 2 {
		return "already checked in"
	}
	if enrollStatus != 1 {
		return "invalid status"
	}
	return ""
}
