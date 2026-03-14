package logic

import (
	"context"
	"errors"

	"pallink/activity/activity"
	"pallink/activity/internal/dao"
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

	if _, _, _, err := dao.GetActivityEnrollInfoForUpdate(l.ctx, tx, in.ActivityId); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &activity.EnrollActivityResponse{Success: false, Message: "activity not found"}, nil
		}
		return nil, err
	}

	status, exists, err := dao.GetEnrollmentStatusForUpdate(l.ctx, tx, in.ActivityId, in.UserId)
	if err != nil {
		return nil, err
	}
	if !exists {
		return &activity.EnrollActivityResponse{Success: false, Message: "not enrolled"}, nil
	}

	if status == 3 {
		return &activity.EnrollActivityResponse{Success: false, Message: "already canceled"}, nil
	}

	if err := dao.UpdateEnrollmentStatus(l.ctx, tx, in.ActivityId, in.UserId, 3, nil, nil); err != nil {
		return nil, err
	}

	if err := dao.UpdateActivityPeople(l.ctx, tx, in.ActivityId, -1); err != nil {
		return nil, err
	}

	if err := tx.Commit(l.ctx); err != nil {
		return nil, err
	}

	return &activity.EnrollActivityResponse{Success: true, Message: "ok"}, nil
}
