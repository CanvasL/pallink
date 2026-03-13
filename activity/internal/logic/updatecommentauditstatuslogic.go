package logic

import (
	"context"

	"pallink/activity/activity"
	"pallink/activity/internal/dao"
	"pallink/activity/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateCommentAuditStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateCommentAuditStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateCommentAuditStatusLogic {
	return &UpdateCommentAuditStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateCommentAuditStatusLogic) UpdateCommentAuditStatus(in *activity.UpdateCommentAuditStatusRequest) (*activity.UpdateCommentAuditStatusResponse, error) {
	ok, err := dao.UpdateCommentAuditStatus(l.ctx, l.svcCtx.DB, in.CommentId, in.AuditStatus)
	if err != nil {
		return nil, err
	}
	if !ok {
		return &activity.UpdateCommentAuditStatusResponse{Success: false, Message: "comment not found"}, nil
	}
	return &activity.UpdateCommentAuditStatusResponse{Success: true, Message: "ok"}, nil
}
