package logic

import (
	"context"
	"errors"

	"pallink/notification/internal/dao"
	"pallink/notification/internal/svc"
	"pallink/notification/notification"

	"github.com/zeromicro/go-zero/core/logx"
)

type MarkReadLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMarkReadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MarkReadLogic {
	return &MarkReadLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *MarkReadLogic) MarkRead(in *notification.MarkReadRequest) (*notification.MarkReadResponse, error) {
	if in.UserId == 0 {
		return nil, errors.New("user_id required")
	}
	if err := dao.MarkRead(l.ctx, l.svcCtx.DB, in.UserId, in.NotificationId); err != nil {
		return nil, err
	}
	return &notification.MarkReadResponse{Success: true, Message: "ok"}, nil
}
