// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"pallink/gateway/internal/svc"
	"pallink/gateway/internal/types"
	"pallink/user/userclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.LoginRequest) (resp *types.LoginResponse, err error) {
	rpcResp, err := l.svcCtx.UserRpc.Login(l.ctx, &userclient.LoginRequest{
		Mobile:   req.Mobile,
		Password: req.Password,
	})
	if err != nil {
		return nil, err
	}

	return &types.LoginResponse{
		UserId:   rpcResp.UserId,
		Token:    rpcResp.Token,
		UserInfo: toUserInfo(rpcResp.UserInfo),
	}, nil
}
