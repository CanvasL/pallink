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

	var (
		maxPeople     int32
		currentPeople int32
		status        int32
	)
	err = tx.QueryRow(
		l.ctx,
		`SELECT max_people, current_people, status FROM activity WHERE id=$1`,
		in.ActivityId,
	).Scan(&maxPeople, &currentPeople, &status)
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

	var existingStatus int32
	err = tx.QueryRow(
		l.ctx,
		`SELECT status FROM enrollment WHERE activity_id=$1 AND user_id=$2`,
		in.ActivityId, in.UserId,
	).Scan(&existingStatus)
	if err == nil {
		if existingStatus == 1 || existingStatus == 2 {
			return &activity.EnrollActivityResponse{Success: false, Message: "already enrolled"}, nil
		}
		_, err = tx.Exec(
			l.ctx,
			`UPDATE enrollment SET status=1, enroll_time=$1, checkin_time=NULL WHERE activity_id=$2 AND user_id=$3`,
			time.Now(), in.ActivityId, in.UserId,
		)
		if err != nil {
			return nil, err
		}
	} else if errors.Is(err, pgx.ErrNoRows) {
		_, err = tx.Exec(
			l.ctx,
			`INSERT INTO enrollment (activity_id, user_id, status, enroll_time) VALUES ($1, $2, 1, $3)`,
			in.ActivityId, in.UserId, time.Now(),
		)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}

	_, err = tx.Exec(
		l.ctx,
		`UPDATE activity SET current_people = current_people + 1 WHERE id=$1`,
		in.ActivityId,
	)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(l.ctx); err != nil {
		return nil, err
	}

	return &activity.EnrollActivityResponse{Success: true, Message: "ok"}, nil
}
