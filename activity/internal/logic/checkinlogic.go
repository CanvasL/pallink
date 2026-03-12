package logic

import (
	"context"
	"errors"
	"time"

	"pallink/activity/activity"
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

	var status int32
	err := l.svcCtx.DB.QueryRow(
		l.ctx,
		`SELECT status FROM enrollment WHERE activity_id=$1 AND user_id=$2`,
		in.ActivityId, in.UserId,
	).Scan(&status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &activity.EnrollActivityResponse{Success: false, Message: "not enrolled"}, nil
		}
		return nil, err
	}

	if status == 2 {
		return &activity.EnrollActivityResponse{Success: false, Message: "already checked in"}, nil
	}
	if status != 1 {
		return &activity.EnrollActivityResponse{Success: false, Message: "invalid status"}, nil
	}

	_, err = l.svcCtx.DB.Exec(
		l.ctx,
		`UPDATE enrollment SET status=2, checkin_time=$1 WHERE activity_id=$2 AND user_id=$3`,
		time.Now(), in.ActivityId, in.UserId,
	)
	if err != nil {
		return nil, err
	}

	return &activity.EnrollActivityResponse{Success: true, Message: "ok"}, nil
}
