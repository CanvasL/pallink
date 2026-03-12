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

type CreateActivityLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateActivityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateActivityLogic {
	return &CreateActivityLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateActivityLogic) CreateActivity(req *types.CreateActivityReq) (resp *types.CreateActivityResp, err error) {
	userID, ok := auth.GetUserIDFromCtx(l.ctx)
	if !ok || userID == 0 {
		return nil, errors.New("unauthorized")
	}
	if req.Title == "" || req.Location == "" {
		return nil, errors.New("title/location required")
	}
	if req.StartTime == 0 || req.EndTime == 0 {
		return nil, errors.New("start_time/end_time required")
	}
	start := time.Unix(req.StartTime, 0)
	end := time.Unix(req.EndTime, 0)
	if !end.After(start) {
		return nil, errors.New("end_time must be after start_time")
	}

	rpcResp, err := l.svcCtx.ActivityRpc.CreateActivity(l.ctx, &activityclient.CreateActivityRequest{
		CreatorId:   userID,
		Title:       req.Title,
		Description: req.Description,
		Location:    req.Location,
		StartTime:   timestamppb.New(start),
		EndTime:     timestamppb.New(end),
		MaxPeople:   req.MaxPeople,
	})
	if err != nil {
		return nil, err
	}

	info := toActivityInfo(rpcResp)
	return &types.CreateActivityResp{Activity: info}, nil
}
