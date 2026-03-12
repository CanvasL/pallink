// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"pallink/activity/activityclient"
	"pallink/gateway/internal/svc"
	"pallink/gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetParticipantsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetParticipantsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetParticipantsLogic {
	return &GetParticipantsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetParticipantsLogic) GetParticipants(req *types.GetParticipantsReq) (resp *types.GetParticipantsResp, err error) {
	page := req.Page
	pageSize := req.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	rpcResp, err := l.svcCtx.ActivityRpc.GetParticipants(l.ctx, &activityclient.GetParticipantsRequest{
		ActivityId: req.Id,
		Page:       page,
		PageSize:   pageSize,
	})
	if err != nil {
		return nil, err
	}

	list := make([]types.ParticipantInfo, 0, len(rpcResp.Participants))
	for _, p := range rpcResp.Participants {
		list = append(list, types.ParticipantInfo{
			UserId:      p.UserId,
			Nickname:    p.Nickname,
			Avatar:      p.Avatar,
			EnrollTime:  tsToUnix(p.EnrollTime),
			CheckinTime: tsToUnix(p.CheckinTime),
			Status:      p.Status,
		})
	}

	return &types.GetParticipantsResp{
		List:     list,
		Total:    rpcResp.Total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}
