// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"pallink/activity/api/internal/svc"
	"pallink/activity/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CancelEnrollLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCancelEnrollLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelEnrollLogic {
	return &CancelEnrollLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CancelEnrollLogic) CancelEnroll(req *types.CancelEnrollReq) (resp *types.CancelEnrollResp, err error) {
	// todo: add your logic here and delete this line

	return
}
