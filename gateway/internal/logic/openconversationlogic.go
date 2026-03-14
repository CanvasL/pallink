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

type OpenConversationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOpenConversationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OpenConversationLogic {
	return &OpenConversationLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *OpenConversationLogic) OpenConversation(req *types.OpenConversationReq) (*types.OpenConversationResp, error) {
	userID, ok := auth.GetUserIDFromCtx(l.ctx)
	if !ok || userID == 0 {
		return nil, errors.New("unauthorized")
	}
	if req.PeerUserId == 0 {
		return nil, errors.New("peer_user_id required")
	}

	rpcResp, err := l.svcCtx.ImRpc.OpenConversation(l.ctx, &imclient.OpenConversationRequest{
		UserId:     userID,
		PeerUserId: req.PeerUserId,
	})
	if err != nil {
		return nil, err
	}
	return &types.OpenConversationResp{
		Conversation: toConversationInfo(rpcResp),
	}, nil
}
