package logic

import (
	"context"

	"pallink/activity/rpc/activity"
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
	return queryActivityDetail(l.ctx, l.svcCtx.DB, in.Id, in.ViewerUserId)
}
