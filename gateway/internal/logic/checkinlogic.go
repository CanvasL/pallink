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

type CheckInLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCheckInLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckInLogic {
	return &CheckInLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CheckInLogic) CheckIn(req *types.CheckInReq) (resp *types.CheckInResp, err error) {
	userID, ok := auth.GetUserIDFromCtx(l.ctx)
	if !ok || userID == 0 {
		return nil, errors.New("unauthorized")
	}
	if req.Id == 0 {
		return nil, errors.New("id required")
	}

	rpcResp, err := l.svcCtx.ActivityRpc.CheckIn(l.ctx, &activityclient.CheckInRequest{
		ActivityId: req.Id,
		UserId:     userID,
		Code:       req.Code,
	})
	if err != nil {
		return nil, err
	}

	return &types.CheckInResp{Success: rpcResp.Success, Message: rpcResp.Message}, nil
}
