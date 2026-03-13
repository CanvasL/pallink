package logic

import (
	"context"
	"errors"
	"strings"

	"pallink/common/auth"
	"pallink/user/internal/dao"
	"pallink/user/internal/svc"
	"pallink/user/user"

	"github.com/jackc/pgx/v5"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
)

type LoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *LoginLogic) Login(in *user.LoginRequest) (*user.LoginResponse, error) {
	mobile := strings.TrimSpace(in.Mobile)
	password := strings.TrimSpace(in.Password)
	if mobile == "" || password == "" {
		return nil, errors.New("mobile/password required")
	}

	userID, passwordHash, nickname, avatar, err := dao.GetLoginInfo(l.ctx, l.svcCtx.DB, mobile)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("invalid mobile or password")
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid mobile or password")
	}

	token, err := auth.GenerateToken(l.svcCtx.Config.Jwt.AccessSecret, l.svcCtx.Config.Jwt.AccessExpire, userID)
	if err != nil {
		return nil, err
	}

	return &user.LoginResponse{
		UserId: userID,
		Token:  token,
		UserInfo: &user.UserInfo{
			Id:       userID,
			Mobile:   mobile,
			Nickname: nickname,
			Avatar:   avatar,
		},
	}, nil
}
