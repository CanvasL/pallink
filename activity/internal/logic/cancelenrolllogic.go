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

type CancelEnrollLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCancelEnrollLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelEnrollLogic {
	return &CancelEnrollLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CancelEnrollLogic) CancelEnroll(in *activity.CancelEnrollRequest) (*activity.EnrollActivityResponse, error) {
	if in.ActivityId == 0 || in.UserId == 0 {
		return &activity.EnrollActivityResponse{Success: false, Message: "activity_id/user_id required"}, nil
	}

	tx, err := l.svcCtx.DB.Begin(l.ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(l.ctx)

	var status int32
	err = tx.QueryRow(
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

	if status == 3 {
		return &activity.EnrollActivityResponse{Success: false, Message: "already canceled"}, nil
	}

	_, err = tx.Exec(
		l.ctx,
		`UPDATE enrollment SET status=3, updated_at=$1 WHERE activity_id=$2 AND user_id=$3`,
		time.Now(), in.ActivityId, in.UserId,
	)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(
		l.ctx,
		`UPDATE activity SET current_people = GREATEST(current_people - 1, 0) WHERE id=$1`,
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
