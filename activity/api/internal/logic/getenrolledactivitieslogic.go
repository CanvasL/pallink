// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"pallink/activity/api/internal/svc"
	"pallink/activity/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetEnrolledActivitiesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetEnrolledActivitiesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetEnrolledActivitiesLogic {
	return &GetEnrolledActivitiesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetEnrolledActivitiesLogic) GetEnrolledActivities(req *types.GetEnrolledActivitiesReq) (resp *types.GetEnrolledActivitiesResp, err error) {
	// todo: add your logic here and delete this line

	return
}
