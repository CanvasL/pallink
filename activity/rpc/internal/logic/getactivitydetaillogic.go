package logic

import (
	"context"

	"pallink/activity"
	"pallink/activity/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetActivityDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetActivityDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetActivityDetailLogic {
	return &GetActivityDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetActivityDetailLogic) GetActivityDetail(in *activity.GetActivityDetailRequest) (*activity.ActivityInfo, error) {
	// todo: add your logic here and delete this line

	return &activity.ActivityInfo{}, nil
}
