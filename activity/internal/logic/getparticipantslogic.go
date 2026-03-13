package logic

import (
	"context"
	"errors"

	"pallink/activity/activity"
	"pallink/activity/internal/dao"
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

	list, total, err := dao.QueryParticipants(l.ctx, l.svcCtx.DB, in.ActivityId, in.Page, in.PageSize)
	if err != nil {
		return nil, err
	}
	if err := hydrateParticipants(l.ctx, l.svcCtx.UserRpc, list); err != nil {
		return nil, err
	}

	return &activity.GetParticipantsResponse{
		Participants: list,
		Total:        total,
	}, nil
}
