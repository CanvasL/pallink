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
	userID, _ := auth.GetUserIDFromCtx(l.ctx)

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
		ViewerUserId: userID,
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
