package logic

import (
	"context"

	"pallink/user/internal/dao"
	"pallink/user/internal/svc"
	"pallink/user/user"

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

func (l *UpdateAuditStatusLogic) UpdateAuditStatus(in *user.UpdateUserAuditStatusRequest) (*user.UpdateUserAuditStatusResponse, error) {
	if in.UserId == 0 {
		return &user.UpdateUserAuditStatusResponse{Success: false, Message: "user_id required"}, nil
	}
	if in.AuditStatus < 0 || in.AuditStatus > 2 {
		return &user.UpdateUserAuditStatusResponse{Success: false, Message: "invalid audit_status"}, nil
	}

	ok, err := dao.UpdateAuditStatus(l.ctx, l.svcCtx.DB, in.UserId, in.AuditStatus)
	if err != nil {
		return nil, err
	}
	if !ok {
		return &user.UpdateUserAuditStatusResponse{Success: false, Message: "user not found"}, nil
	}
	return &user.UpdateUserAuditStatusResponse{Success: true, Message: "ok"}, nil
}
