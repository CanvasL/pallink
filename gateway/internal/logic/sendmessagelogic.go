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

type SendMessageLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSendMessageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendMessageLogic {
	return &SendMessageLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SendMessageLogic) SendMessage(req *types.SendMessageReq) (*types.SendMessageResp, error) {
	userID, ok := auth.GetUserIDFromCtx(l.ctx)
	if !ok || userID == 0 {
		return nil, errors.New("unauthorized")
	}
	if req.PeerUserId == 0 {
		return nil, errors.New("peer_user_id required")
	}
	if req.Content == "" {
		return nil, errors.New("content required")
	}

	rpcResp, err := l.svcCtx.ImRpc.SendMessage(l.ctx, &imclient.SendMessageRequest{
		SenderId:   userID,
		PeerUserId: req.PeerUserId,
		Content:    req.Content,
	})
	if err != nil {
		return nil, err
	}
	return &types.SendMessageResp{
		Message: toMessageInfo(rpcResp),
	}, nil
}
