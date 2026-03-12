package logic

import (
	"context"

	"pallink/activity"
	"pallink/activity/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type EnrollActivityLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewEnrollActivityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *EnrollActivityLogic {
	return &EnrollActivityLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *EnrollActivityLogic) EnrollActivity(in *activity.EnrollActivityRequest) (*activity.EnrollActivityResponse, error) {
	// todo: add your logic here and delete this line

	return &activity.EnrollActivityResponse{}, nil
}
