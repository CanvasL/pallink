package logic

import (
	"context"
	"errors"

	"pallink/activity/activity"
	"pallink/activity/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetParticipantsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetParticipantsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetParticipantsLogic {
	return &GetParticipantsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetParticipantsLogic) GetParticipants(in *activity.GetParticipantsRequest) (*activity.GetParticipantsResponse, error) {
	if in.ActivityId == 0 {
		return nil, errors.New("activity_id required")
	}

	list, total, err := queryParticipants(l.ctx, l.svcCtx.DB, in.ActivityId, in.Page, in.PageSize)
	if err != nil {
		return nil, err
	}

	return &activity.GetParticipantsResponse{
		Participants: list,
		Total:        total,
	}, nil
}
