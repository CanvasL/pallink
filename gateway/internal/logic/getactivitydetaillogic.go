// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"pallink/activity/activityclient"
	"pallink/common/auth"
	"pallink/gateway/internal/svc"
	"pallink/gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetActivityDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetActivityDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetActivityDetailLogic {
	return &GetActivityDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetActivityDetailLogic) GetActivityDetail(req *types.GetActivityDetailReq) (resp *types.GetActivityDetailResp, err error) {
	userID, _ := auth.GetUserIDFromCtx(l.ctx)

	rpcResp, err := l.svcCtx.ActivityRpc.GetActivityDetail(l.ctx, &activityclient.GetActivityDetailRequest{
		Id:           req.Id,
		ViewerUserId: userID,
	})
	if err != nil {
		return nil, err
	}

	info := toActivityInfo(rpcResp)
	return &types.GetActivityDetailResp{Activity: info}, nil
}
