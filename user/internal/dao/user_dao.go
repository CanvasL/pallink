package dao

import (
	"context"
	"errors"

	"pallink/user/internal/dao/sqlc"
	"pallink/user/user"

	"github.com/jackc/pgx/v5"
)

func GetUserIDByMobile(ctx context.Context, db sqlc.DBTX, mobile string) (uint64, bool, error) {
	q := sqlc.New(db)
	id, err := q.GetUserIDByMobile(ctx, mobile)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, false, nil
		}
		return 0, false, err
	}
	return uint64(id), true, nil
}

func InsertUser(ctx context.Context, db sqlc.DBTX, mobile, passwordHash, nickname, avatar string) (uint64, error) {
	q := sqlc.New(db)
	id, err := q.InsertUser(ctx, sqlc.InsertUserParams{
		Mobile:       mobile,
		PasswordHash: passwordHash,
		Nickname:     nickname,
		Avatar:       avatar,
	})
	if err != nil {
		return 0, err
	}
	return uint64(id), nil
}

func GetLoginInfo(ctx context.Context, db sqlc.DBTX, mobile string) (uint64, string, string, string, error) {
	q := sqlc.New(db)
	row, err := q.GetLoginInfo(ctx, mobile)
	if err != nil {
		return 0, "", "", "", err
	}
	return uint64(row.ID), row.PasswordHash, row.Nickname, row.Avatar, nil
}

func GetUserInfo(ctx context.Context, db sqlc.DBTX, userID uint64) (*user.UserInfo, error) {
	q := sqlc.New(db)
	row, err := q.GetUserInfo(ctx, int64(userID))
	if err != nil {
		return nil, err
	}

	return &user.UserInfo{
		Id:          uint64(row.ID),
		Mobile:      row.Mobile,
		Nickname:    row.Nickname,
		Avatar:      row.Avatar,
		AuditStatus: int32(row.AuditStatus),
	}, nil
}

func UpdateUserInfo(ctx context.Context, db sqlc.DBTX, userID uint64, nickname, avatar string) (bool, error) {
	q := sqlc.New(db)
	affected, err := q.UpdateUserInfo(ctx, sqlc.UpdateUserInfoParams{
		ID:       int64(userID),
		Nickname: nickname,
		Avatar:   avatar,
	})
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}

func UpdateAuditStatus(ctx context.Context, db sqlc.DBTX, userID uint64, auditStatus int32) (bool, error) {
	q := sqlc.New(db)
	affected, err := q.UpdateUserAuditStatus(ctx, sqlc.UpdateUserAuditStatusParams{
		ID:          int64(userID),
		AuditStatus: int16(auditStatus),
	})
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}
