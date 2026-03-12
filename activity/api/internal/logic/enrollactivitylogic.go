// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"pallink/activity/api/internal/svc"
	"pallink/activity/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type EnrollActivityLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewEnrollActivityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *EnrollActivityLogic {
	return &EnrollActivityLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *EnrollActivityLogic) EnrollActivity(req *types.EnrollActivityReq) (resp *types.EnrollActivityResp, err error) {
	// todo: add your logic here and delete this line

	return
}
