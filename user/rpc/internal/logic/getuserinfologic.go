package logic

import (
	"context"
	"errors"

	"pallink/user/rpc/internal/svc"
	"pallink/user/rpc/user"

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

	var (
		mobile   string
		nickname string
		avatar   string
	)
	err := l.svcCtx.DB.QueryRow(
		l.ctx,
		`SELECT mobile, nickname, avatar FROM "user" WHERE id=$1`,
		in.UserId,
	).Scan(&mobile, &nickname, &avatar)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &user.UserInfo{
		Id:       in.UserId,
		Mobile:   mobile,
		Nickname: nickname,
		Avatar:   avatar,
	}, nil
}
