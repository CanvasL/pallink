package logic

import (
	"context"
	"errors"
	"strings"

	"pallink/common/auth"
	"pallink/common/mq"
	"pallink/user/internal/svc"
	"pallink/user/user"

	"github.com/jackc/pgx/v5"
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

	var existingID uint64
	err := l.svcCtx.DB.QueryRow(l.ctx, `SELECT id FROM "user" WHERE mobile=$1`, mobile).Scan(&existingID)
	if err == nil {
		return nil, errors.New("mobile already registered")
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	var userID uint64
	err = l.svcCtx.DB.QueryRow(
		l.ctx,
		`INSERT INTO "user" (mobile, password_hash, nickname, avatar) VALUES ($1, $2, $3, $4) RETURNING id`,
		mobile, string(hashed), nickname, "",
	).Scan(&userID)
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
