package logic

import (
	"context"
	"errors"

	"pallink/common/auth"
	"pallink/gateway/internal/svc"
	"pallink/gateway/internal/types"
	"pallink/im/imclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetConversationListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetConversationListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetConversationListLogic {
	return &GetConversationListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetConversationListLogic) GetConversationList(req *types.GetConversationListReq) (*types.GetConversationListResp, error) {
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

	rpcResp, err := l.svcCtx.ImRpc.GetConversationList(l.ctx, &imclient.GetConversationListRequest{
		UserId:   userID,
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		return nil, err
	}

	list := make([]types.ConversationInfo, 0, len(rpcResp.List))
	for _, item := range rpcResp.List {
		list = append(list, toConversationInfo(item))
	}
	return &types.GetConversationListResp{
		List:     list,
		Total:    rpcResp.Total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}
