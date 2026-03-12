// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"pallink/user/api/internal/svc"
	"pallink/user/api/internal/types"
	"pallink/user/rpc/userclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type RegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegisterLogic) Register(req *types.RegisterRequest) (resp *types.RegisterResponse, err error) {
	rpcResp, err := l.svcCtx.UserRpc.Register(l.ctx, &userclient.RegisterRequest{
		Mobile:     req.Mobile,
		Password:   req.Password,
		Nickname:   req.Nickname,
		VerifyCode: req.VerifyCode,
	})
	if err != nil {
		return nil, err
	}

	return &types.RegisterResponse{
		UserId: rpcResp.UserId,
		Token:  rpcResp.Token,
	}, nil
}
