// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"pallink/activity/api/internal/svc"
	"pallink/activity/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMyActivitiesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetMyActivitiesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMyActivitiesLogic {
	return &GetMyActivitiesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMyActivitiesLogic) GetMyActivities(req *types.GetMyActivitiesReq) (resp *types.GetMyActivitiesResp, err error) {
	// todo: add your logic here and delete this line

	return
}
