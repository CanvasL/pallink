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

type GetMessagesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetMessagesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMessagesLogic {
	return &GetMessagesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMessagesLogic) GetMessages(req *types.GetMessagesReq) (*types.GetMessagesResp, error) {
	userID, ok := auth.GetUserIDFromCtx(l.ctx)
	if !ok || userID == 0 {
		return nil, errors.New("unauthorized")
	}
	if req.ConversationId == 0 {
		return nil, errors.New("conversation_id required")
	}

	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}

	rpcResp, err := l.svcCtx.ImRpc.GetMessages(l.ctx, &imclient.GetMessagesRequest{
		UserId:          userID,
		ConversationId:  req.ConversationId,
		BeforeMessageId: req.BeforeMessageId,
		PageSize:        pageSize,
	})
	if err != nil {
		return nil, err
	}

	list := make([]types.MessageInfo, 0, len(rpcResp.List))
	for _, item := range rpcResp.List {
		list = append(list, toMessageInfo(item))
	}
	return &types.GetMessagesResp{
		List:     list,
		HasMore:  rpcResp.HasMore,
		PageSize: pageSize,
	}, nil
}
