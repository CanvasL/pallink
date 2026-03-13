// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"pallink/activity/activityclient"
	"pallink/gateway/internal/svc"
	"pallink/gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPublicCommentsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetPublicCommentsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPublicCommentsLogic {
	return &GetPublicCommentsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPublicCommentsLogic) GetPublicComments(req *types.GetCommentsReq) (resp *types.GetCommentsResp, err error) {
	page := req.Page
	pageSize := req.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	rpcResp, err := l.svcCtx.ActivityRpc.GetComments(l.ctx, &activityclient.GetCommentsRequest{
		ActivityId:   req.ActivityId,
		Page:         page,
		PageSize:     pageSize,
		ViewerUserId: 0,
		ParentId:     req.ParentId,
	})
	if err != nil {
		return nil, err
	}

	list := make([]types.CommentInfo, 0, len(rpcResp.Comments))
	for _, item := range rpcResp.Comments {
		list = append(list, toCommentInfo(item))
	}

	return &types.GetCommentsResp{
		List:     list,
		Total:    rpcResp.Total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}
