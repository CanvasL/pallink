package logic

import (
	"context"
	"errors"
	"strings"

	"pallink/common/mq"
	"pallink/user/internal/svc"
	"pallink/user/user"

	sq "github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateUserInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserInfoLogic {
	return &UpdateUserInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateUserInfoLogic) UpdateUserInfo(in *user.UpdateUserInfoRequest) (*user.UserInfo, error) {
	if in.UserId == 0 {
		return nil, errors.New("user_id required")
	}

	if in.Nickname == "" && in.Avatar == "" {
		return nil, errors.New("no fields to update")
	}

	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Update(`"user"`).
		Set("audit_status", 0).
		Set("updated_at", sq.Expr("now()")).
		Where(sq.Eq{"id": in.UserId})

	if in.Nickname != "" {
		builder = builder.Set("nickname", strings.TrimSpace(in.Nickname))
	}
	if in.Avatar != "" {
		builder = builder.Set("avatar", strings.TrimSpace(in.Avatar))
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}
	cmd, err := l.svcCtx.DB.Exec(l.ctx, query, args...)
	if err != nil {
		return nil, err
	}
	if cmd.RowsAffected() == 0 {
		return nil, errors.New("user not found")
	}

	if err := l.svcCtx.MQ.PublishJSON(l.ctx, mq.AuditMessage{Type: "user", ID: in.UserId}); err != nil {
		return nil, err
	}

	return NewGetUserInfoLogic(l.ctx, l.svcCtx).GetUserInfo(&user.GetUserInfoRequest{UserId: in.UserId})
}
