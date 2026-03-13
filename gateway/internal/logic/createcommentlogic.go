// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"
	"errors"

	"pallink/activity/activityclient"
	"pallink/common/auth"
	"pallink/gateway/internal/svc"
	"pallink/gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateCommentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateCommentLogic {
	return &CreateCommentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateCommentLogic) CreateComment(req *types.CreateCommentReq) (resp *types.CreateCommentResp, err error) {
	userID, ok := auth.GetUserIDFromCtx(l.ctx)
	if !ok || userID == 0 {
		return nil, errors.New("unauthorized")
	}
	if req.ActivityId == 0 {
		return nil, errors.New("activity_id required")
	}
	if req.Content == "" {
		return nil, errors.New("content required")
	}

	rpcResp, err := l.svcCtx.ActivityRpc.CreateComment(l.ctx, &activityclient.CreateCommentRequest{
		ActivityId: req.ActivityId,
		UserId:     userID,
		Content:    req.Content,
	})
	if err != nil {
		return nil, err
	}
	return &types.CreateCommentResp{CommentId: rpcResp.CommentId}, nil
}
