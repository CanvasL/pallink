package logic

import (
	"context"
	"errors"

	"pallink/notification/internal/dao"
	"pallink/notification/internal/svc"
	"pallink/notification/notification"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetNotificationsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetNotificationsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetNotificationsLogic {
	return &GetNotificationsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetNotificationsLogic) GetNotifications(in *notification.GetNotificationsRequest) (*notification.GetNotificationsResponse, error) {
	if in.UserId == 0 {
		return nil, errors.New("user_id required")
	}

	list, total, err := dao.QueryNotifications(l.ctx, l.svcCtx.DB, dao.ListFilter{
		UserID:     in.UserId,
		UnreadOnly: in.UnreadOnly,
		Page:       in.Page,
		PageSize:   in.PageSize,
	})
	if err != nil {
		return nil, err
	}

	return &notification.GetNotificationsResponse{List: list, Total: total}, nil
}
