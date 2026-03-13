package logic

import (
	"context"
	"errors"

	"pallink/activity/activity"
	"pallink/activity/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetActivityDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetActivityDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetActivityDetailLogic {
	return &GetActivityDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetActivityDetailLogic) GetActivityDetail(in *activity.GetActivityDetailRequest) (*activity.ActivityInfo, error) {
	info, err := queryActivityDetail(l.ctx, l.svcCtx.DB, in.Id, in.ViewerUserId)
	if err != nil {
		return nil, err
	}
	if in.ViewerUserId == 0 && info.AuditStatus != 1 {
		return nil, errors.New("activity not approved")
	}
	if err := hydrateActivityUsers(l.ctx, l.svcCtx.UserRpc, []*activity.ActivityInfo{info}); err != nil {
		return nil, err
	}
	return info, nil
}
