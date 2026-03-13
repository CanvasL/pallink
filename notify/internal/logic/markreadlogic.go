package logic

import (
	"context"
	"errors"

	"pallink/notify/internal/dao"
	"pallink/notify/internal/svc"
	"pallink/notify/notify"

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

func (l *MarkReadLogic) MarkRead(in *notify.MarkReadRequest) (*notify.MarkReadResponse, error) {
	if in.UserId == 0 {
		return nil, errors.New("user_id required")
	}
	if err := dao.MarkRead(l.ctx, l.svcCtx.DB, in.UserId, in.NotificationId); err != nil {
		return nil, err
	}
	return &notify.MarkReadResponse{Success: true, Message: "ok"}, nil
}
