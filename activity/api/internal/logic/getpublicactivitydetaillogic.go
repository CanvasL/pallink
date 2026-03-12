// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"pallink/activity/api/internal/svc"
	"pallink/activity/api/internal/types"
	"pallink/activity/rpc/activityclient"

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
	rpcResp, err := l.svcCtx.ActivityRpc.GetActivityDetail(l.ctx, &activityclient.GetActivityDetailRequest{
		Id:           req.Id,
		ViewerUserId: 0,
	})
	if err != nil {
		return nil, err
	}

	return &types.GetActivityDetailResp{Activity: toActivityInfo(rpcResp)}, nil
}
