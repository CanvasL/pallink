// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"
	"errors"

	"pallink/activity/activityclient"
	"pallink/common/auth"
	"pallink/gateway/internal/svc"
	"pallink/gateway/internal/types"

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
	userID, ok := auth.GetUserIDFromCtx(l.ctx)
	if !ok || userID == 0 {
		return nil, errors.New("unauthorized")
	}
	if req.Id == 0 {
		return nil, errors.New("id required")
	}

	rpcResp, err := l.svcCtx.ActivityRpc.CancelEnroll(l.ctx, &activityclient.CancelEnrollRequest{
		ActivityId: req.Id,
		UserId:     userID,
	})
	if err != nil {
		return nil, err
	}

	return &types.CancelEnrollResp{Success: rpcResp.Success, Message: rpcResp.Message}, nil
}
