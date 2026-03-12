// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"
	"errors"
	"time"

	"pallink/activity/activityclient"
	"pallink/common/auth"
	"pallink/gateway/internal/svc"
	"pallink/gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UpdateActivityLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateActivityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateActivityLogic {
	return &UpdateActivityLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateActivityLogic) UpdateActivity(req *types.UpdateActivityReq) (resp *types.UpdateActivityResp, err error) {
	userID, ok := auth.GetUserIDFromCtx(l.ctx)
	if !ok || userID == 0 {
		return nil, errors.New("unauthorized")
	}

	maxPeople := int32(-1)
	if req.MaxPeople != nil {
		maxPeople = *req.MaxPeople
	}
	status := int32(-1)
	if req.Status != nil {
		status = *req.Status
	}

	updateReq := &activityclient.UpdateActivityRequest{
		Id:        req.Id,
		CreatorId: userID,
		MaxPeople: maxPeople,
		Status:    status,
	}
	if req.Title != nil {
		updateReq.Title = *req.Title
	}
	if req.Description != nil {
		updateReq.Description = *req.Description
	}
	if req.Location != nil {
		updateReq.Location = *req.Location
	}
	if req.StartTime != nil {
		updateReq.StartTime = timestamppb.New(time.Unix(*req.StartTime, 0))
	}
	if req.EndTime != nil {
		updateReq.EndTime = timestamppb.New(time.Unix(*req.EndTime, 0))
	}

	rpcResp, err := l.svcCtx.ActivityRpc.UpdateActivity(l.ctx, updateReq)
	if err != nil {
		return nil, err
	}

	info := toActivityInfo(rpcResp)
	return &types.UpdateActivityResp{Activity: info}, nil
}
