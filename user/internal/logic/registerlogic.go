package logic

import (
	"context"
	"errors"
	"strings"

	"pallink/common/auth"
	"pallink/common/mq"
	"pallink/user/internal/dao"
	"pallink/user/internal/svc"
	"pallink/user/user"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
)

type RegisterLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RegisterLogic) Register(in *user.RegisterRequest) (*user.RegisterResponse, error) {
	mobile := strings.TrimSpace(in.Mobile)
	password := strings.TrimSpace(in.Password)
	nickname := strings.TrimSpace(in.Nickname)

	if mobile == "" || password == "" || nickname == "" {
		return nil, errors.New("mobile/password/nickname required")
	}

	if _, exists, err := dao.GetUserIDByMobile(l.ctx, l.svcCtx.DB, mobile); err != nil {
		return nil, err
	} else if exists {
		return nil, errors.New("mobile already registered")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	userID, err := dao.InsertUser(l.ctx, l.svcCtx.DB, mobile, string(hashed), nickname, "")
	if err != nil {
		return nil, err
	}
	if err := l.svcCtx.MQ.PublishJSON(l.ctx, mq.AuditMessage{Type: "user", ID: userID}); err != nil {
		return nil, err
	}

	token, err := auth.GenerateToken(l.svcCtx.Config.Jwt.AccessSecret, l.svcCtx.Config.Jwt.AccessExpire, userID)
	if err != nil {
		return nil, err
	}

	return &user.RegisterResponse{
		UserId: userID,
		Token:  token,
	}, nil
}
