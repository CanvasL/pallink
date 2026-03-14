package logic

import (
	"context"
	"errors"
	"strings"

	"pallink/common/mq"
	"pallink/user/internal/dao"
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

	if in.Nickname == "" && in.Avatar == "" {
		return nil, errors.New("no fields to update")
	}

	nickname := strings.TrimSpace(in.Nickname)
	avatar := strings.TrimSpace(in.Avatar)
	ok, err := dao.UpdateUserInfo(l.ctx, l.svcCtx.DB, in.UserId, nickname, avatar)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("user not found")
	}

	if err := l.svcCtx.MQ.PublishJSON(l.ctx, mq.AuditMessage{Type: "user", ID: in.UserId}); err != nil {
		return nil, err
	}

	return NewGetUserInfoLogic(l.ctx, l.svcCtx).GetUserInfo(&user.GetUserInfoRequest{UserId: in.UserId})
}
