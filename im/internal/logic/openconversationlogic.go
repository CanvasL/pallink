package logic

import (
	"context"
	"errors"

	"pallink/im/im"
	"pallink/im/internal/dao"
	"pallink/im/internal/svc"
	"pallink/user/userclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type OpenConversationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewOpenConversationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OpenConversationLogic {
	return &OpenConversationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *OpenConversationLogic) OpenConversation(in *im.OpenConversationRequest) (*im.ConversationInfo, error) {
	if in.UserId == 0 || in.PeerUserId == 0 {
		return nil, errors.New("user_id/peer_user_id required")
	}
	if in.UserId == in.PeerUserId {
		return nil, errors.New("cannot chat with self")
	}

	if _, err := l.svcCtx.UserRpc.GetUserInfo(l.ctx, &userclient.GetUserInfoRequest{UserId: in.PeerUserId}); err != nil {
		return nil, err
	}

	tx, err := l.svcCtx.DB.Begin(l.ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(l.ctx)

	info, err := dao.GetOrCreateConversation(l.ctx, tx, in.UserId, in.PeerUserId)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(l.ctx); err != nil {
		return nil, err
	}

	if err := hydrateConversations(l.ctx, l.svcCtx.UserRpc, []*im.ConversationInfo{info}); err != nil {
		return nil, err
	}
	return info, nil
}
