package logic

import (
	"context"

	"pallink/activity/activity"
	"pallink/activity/internal/dao"
	"pallink/activity/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateAuditStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateAuditStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateAuditStatusLogic {
	return &UpdateAuditStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateAuditStatusLogic) UpdateAuditStatus(in *activity.UpdateAuditStatusRequest) (*activity.UpdateAuditStatusResponse, error) {
	if in.ActivityId == 0 {
		return &activity.UpdateAuditStatusResponse{Success: false, Message: "activity_id required"}, nil
	}
	if in.AuditStatus < 0 || in.AuditStatus > 2 {
		return &activity.UpdateAuditStatusResponse{Success: false, Message: "invalid audit_status"}, nil
	}

	ok, err := dao.UpdateAuditStatus(l.ctx, l.svcCtx.DB, in.ActivityId, in.AuditStatus)
	if err != nil {
		return nil, err
	}
	if !ok {
		return &activity.UpdateAuditStatusResponse{Success: false, Message: "activity not found"}, nil
	}

	return &activity.UpdateAuditStatusResponse{Success: true, Message: "ok"}, nil
}
