// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"
	"errors"

	"pallink/common/auth"
	"pallink/gateway/internal/svc"
	"pallink/gateway/internal/types"
	"pallink/notify/notifyclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type MarkReadLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMarkReadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MarkReadLogic {
	return &MarkReadLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MarkReadLogic) MarkRead(req *types.MarkReadReq) (resp *types.MarkReadResp, err error) {
	userID, ok := auth.GetUserIDFromCtx(l.ctx)
	if !ok || userID == 0 {
		return nil, errors.New("unauthorized")
	}

	_, err = l.svcCtx.NotifyRpc.MarkRead(l.ctx, &notifyclient.MarkReadRequest{
		UserId:         userID,
		NotificationId: req.NotificationId,
	})
	if err != nil {
		return nil, err
	}

	return &types.MarkReadResp{Success: true, Message: "ok"}, nil
}
