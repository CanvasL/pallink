// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"pallink/activity/api/internal/svc"
	"pallink/activity/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPublicActivityDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetPublicActivityDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPublicActivityDetailLogic {
	return &GetPublicActivityDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPublicActivityDetailLogic) GetPublicActivityDetail(req *types.GetActivityDetailReq) (resp *types.GetActivityDetailResp, err error) {
	// todo: add your logic here and delete this line

	return
}
