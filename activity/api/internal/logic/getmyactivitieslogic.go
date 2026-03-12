// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"
	"errors"

	"pallink/activity/api/internal/svc"
	"pallink/activity/api/internal/types"
	"pallink/activity/rpc/activityclient"
	"pallink/common/auth"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMyActivitiesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetMyActivitiesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMyActivitiesLogic {
	return &GetMyActivitiesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMyActivitiesLogic) GetMyActivities(req *types.GetMyActivitiesReq) (resp *types.GetMyActivitiesResp, err error) {
	userID, ok := auth.GetUserIDFromCtx(l.ctx)
	if !ok || userID == 0 {
		return nil, errors.New("unauthorized")
	}

	page := req.Page
	pageSize := req.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	rpcResp, err := l.svcCtx.ActivityRpc.GetMyActivities(l.ctx, &activityclient.GetMyActivitiesRequest{
		UserId:   userID,
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		return nil, err
	}

	list := make([]types.ActivityBrief, 0, len(rpcResp.Activities))
	for _, item := range rpcResp.Activities {
		list = append(list, toActivityBrief(item))
	}

	return &types.GetMyActivitiesResp{
		List:     list,
		Total:    rpcResp.Total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}
