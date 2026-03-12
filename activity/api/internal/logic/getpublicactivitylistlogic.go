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

type GetPublicActivityListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetPublicActivityListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPublicActivityListLogic {
	return &GetPublicActivityListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPublicActivityListLogic) GetPublicActivityList(req *types.GetActivityListReq) (resp *types.GetActivityListResp, err error) {
	page := req.Page
	pageSize := req.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	rpcResp, err := l.svcCtx.ActivityRpc.GetActivityList(l.ctx, &activityclient.GetActivityListRequest{
		Page:         page,
		PageSize:     pageSize,
		Status:       req.Status,
		Keyword:      req.Keyword,
		ViewerUserId: 0,
	})
	if err != nil {
		return nil, err
	}

	list := make([]types.ActivityBrief, 0, len(rpcResp.Activities))
	for _, item := range rpcResp.Activities {
		list = append(list, toActivityBrief(item))
	}

	return &types.GetActivityListResp{
		List:     list,
		Total:    rpcResp.Total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}
