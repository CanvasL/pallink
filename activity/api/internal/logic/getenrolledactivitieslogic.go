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

type GetEnrolledActivitiesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetEnrolledActivitiesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetEnrolledActivitiesLogic {
	return &GetEnrolledActivitiesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetEnrolledActivitiesLogic) GetEnrolledActivities(req *types.GetEnrolledActivitiesReq) (resp *types.GetEnrolledActivitiesResp, err error) {
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

	rpcResp, err := l.svcCtx.ActivityRpc.GetEnrolledActivities(l.ctx, &activityclient.GetEnrolledActivitiesRequest{
		UserId:   userID,
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		return nil, err
	}

	list := make([]types.ActivityBrief, 0, len(rpcResp.Activities))
	for _, item := range rpcResp.Activities {
		brief := toActivityBrief(item)
		brief.IsEnrolled = true
		list = append(list, brief)
	}

	return &types.GetEnrolledActivitiesResp{
		List:     list,
		Total:    rpcResp.Total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}
