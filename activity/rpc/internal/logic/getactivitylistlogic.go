package logic

import (
	"context"

	"pallink/activity/rpc/activity"
	"pallink/activity/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetActivityListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetActivityListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetActivityListLogic {
	return &GetActivityListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetActivityListLogic) GetActivityList(in *activity.GetActivityListRequest) (*activity.GetActivityListResponse, error) {
	// todo: add your logic here and delete this line

	return &activity.GetActivityListResponse{}, nil
}
