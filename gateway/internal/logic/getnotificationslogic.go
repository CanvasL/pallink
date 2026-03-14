// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"
	"errors"

	"pallink/common/auth"
	"pallink/gateway/internal/svc"
	"pallink/gateway/internal/types"
	"pallink/notification/notificationclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetNotificationsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetNotificationsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetNotificationsLogic {
	return &GetNotificationsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetNotificationsLogic) GetNotifications(req *types.GetNotificationsReq) (resp *types.GetNotificationsResp, err error) {
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

	rpcResp, err := l.svcCtx.NotificationRpc.GetNotifications(l.ctx, &notificationclient.GetNotificationsRequest{
		UserId:     userID,
		Page:       page,
		PageSize:   pageSize,
		UnreadOnly: req.UnreadOnly,
	})
	if err != nil {
		return nil, err
	}

	list := make([]types.NotificationInfo, 0, len(rpcResp.List))
	for _, item := range rpcResp.List {
		if item == nil {
			continue
		}
		readAt := int64(0)
		if item.ReadAt != nil {
			readAt = item.ReadAt.AsTime().Unix()
		}
		list = append(list, types.NotificationInfo{
			Id:         item.Id,
			UserId:     item.UserId,
			ActorId:    item.ActorId,
			Type:       item.Type,
			ActivityId: item.ActivityId,
			CommentId:  item.CommentId,
			ParentId:   item.ParentId,
			Content:    item.Content,
			CreatedAt:  item.CreatedAt.AsTime().Unix(),
			ReadAt:     readAt,
		})
	}

	return &types.GetNotificationsResp{
		List:     list,
		Total:    rpcResp.Total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}
