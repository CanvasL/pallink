package logic

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"pallink/common/mq"
	"pallink/user/internal/svc"
	"pallink/user/user"

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

	sets := make([]string, 0)
	args := make([]any, 0)
	addArg := func(v any) string {
		args = append(args, v)
		return "$" + strconv.Itoa(len(args))
	}

	if in.Nickname != "" {
		sets = append(sets, "nickname="+addArg(strings.TrimSpace(in.Nickname)))
	}
	if in.Avatar != "" {
		sets = append(sets, "avatar="+addArg(strings.TrimSpace(in.Avatar)))
	}
	if len(sets) == 0 {
		return nil, errors.New("no fields to update")
	}

	sets = append(sets, "audit_status=0", "updated_at=now()")
	userArg := addArg(in.UserId)
	query := "UPDATE \"user\" SET " + strings.Join(sets, ", ") + " WHERE id=" + userArg
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
