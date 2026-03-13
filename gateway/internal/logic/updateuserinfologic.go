// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"
	"errors"

	"pallink/common/auth"
	"pallink/gateway/internal/svc"
	"pallink/gateway/internal/types"
	"pallink/user/userclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateUserInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserInfoLogic {
	return &UpdateUserInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateUserInfoLogic) UpdateUserInfo(req *types.UpdateUserRequest) (resp *types.UpdateUserResponse, err error) {
	userID, ok := auth.GetUserIDFromCtx(l.ctx)
	if !ok || userID == 0 {
		return nil, errors.New("unauthorized")
	}

	rpcResp, err := l.svcCtx.UserRpc.UpdateUserInfo(l.ctx, &userclient.UpdateUserInfoRequest{
		UserId:   userID,
		Nickname: req.Nickname,
		Avatar:   req.Avatar,
	})
	if err != nil {
		return nil, err
	}

	info := toUserInfo(rpcResp)
	return &types.UpdateUserResponse{UserInfo: info}, nil
}
