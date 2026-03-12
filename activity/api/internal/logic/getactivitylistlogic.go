// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"pallink/activity/api/internal/svc"
	"pallink/activity/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetActivityListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetActivityListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetActivityListLogic {
	return &GetActivityListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetActivityListLogic) GetActivityList(req *types.GetActivityListReq) (resp *types.GetActivityListResp, err error) {
	// todo: add your logic here and delete this line

	return
}
