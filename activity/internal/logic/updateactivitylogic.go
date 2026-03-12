package logic

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"pallink/activity/activity"
	"pallink/activity/internal/svc"

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

	sets := make([]string, 0)
	args := make([]any, 0)
	addArg := func(v any) string {
		args = append(args, v)
		return fmt.Sprintf("$%d", len(args))
	}

	if in.Title != "" {
		sets = append(sets, "title="+addArg(in.Title))
	}
	if in.Description != "" {
		sets = append(sets, "description="+addArg(in.Description))
	}
	if in.Location != "" {
		sets = append(sets, "location="+addArg(in.Location))
	}
	if in.StartTime != nil {
		sets = append(sets, "start_time="+addArg(in.StartTime.AsTime()))
	}
	if in.EndTime != nil {
		sets = append(sets, "end_time="+addArg(in.EndTime.AsTime()))
	}
	if in.MaxPeople != -1 {
		sets = append(sets, "max_people="+addArg(in.MaxPeople))
	}
	if in.Status != -1 {
		sets = append(sets, "status="+addArg(in.Status))
	}

	if len(sets) == 0 {
		return nil, errors.New("no fields to update")
	}

	if in.StartTime != nil && in.EndTime != nil {
		if !in.EndTime.AsTime().After(in.StartTime.AsTime()) {
			return nil, errors.New("end_time must be after start_time")
		}
	}

	sets = append(sets, "updated_at=now()")
	idArg := addArg(in.Id)
	creatorArg := addArg(in.CreatorId)
	query := "UPDATE activity SET " + strings.Join(sets, ", ") + " WHERE id=" + idArg + " AND creator_id=" + creatorArg

	cmd, err := l.svcCtx.DB.Exec(l.ctx, query, args...)
	if err != nil {
		return nil, err
	}
	if cmd.RowsAffected() == 0 {
		return nil, errors.New("activity not found or forbidden")
	}

	return queryActivityDetail(l.ctx, l.svcCtx.DB, in.Id, in.CreatorId)
}
