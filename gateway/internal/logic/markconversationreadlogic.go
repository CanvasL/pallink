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

type MarkConversationReadLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMarkConversationReadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MarkConversationReadLogic {
	return &MarkConversationReadLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MarkConversationReadLogic) MarkConversationRead(req *types.MarkConversationReadReq) (*types.MarkConversationReadResp, error) {
	userID, ok := auth.GetUserIDFromCtx(l.ctx)
	if !ok || userID == 0 {
		return nil, errors.New("unauthorized")
	}
	if req.ConversationId == 0 {
		return nil, errors.New("conversation_id required")
	}

	rpcResp, err := l.svcCtx.ImRpc.MarkConversationRead(l.ctx, &imclient.MarkConversationReadRequest{
		UserId:         userID,
		ConversationId: req.ConversationId,
		LastReadMsgId:  req.LastReadMsgId,
	})
	if err != nil {
		return nil, err
	}
	return &types.MarkConversationReadResp{
		Success:       rpcResp.Success,
		Message:       rpcResp.Message,
		LastReadMsgId: rpcResp.LastReadMsgId,
	}, nil
}
