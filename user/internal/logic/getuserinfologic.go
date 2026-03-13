package logic

import (
	"context"
	"errors"

	"pallink/user/internal/dao"
	"pallink/user/internal/svc"
	"pallink/user/user"

	"github.com/jackc/pgx/v5"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserInfoLogic {
	return &GetUserInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserInfoLogic) GetUserInfo(in *user.GetUserInfoRequest) (*user.UserInfo, error) {
	if in.UserId == 0 {
		return nil, errors.New("user_id required")
	}

	info, err := dao.GetUserInfo(l.ctx, l.svcCtx.DB, in.UserId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return info, nil
}
