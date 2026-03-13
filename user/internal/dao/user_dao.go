package dao

import (
	"context"
	"errors"

	"pallink/user/user"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func GetUserIDByMobile(ctx context.Context, db *pgxpool.Pool, mobile string) (uint64, bool, error) {
	var id uint64
	err := db.QueryRow(ctx, `SELECT id FROM "user" WHERE mobile=$1`, mobile).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, false, nil
		}
		return 0, false, err
	}
	return id, true, nil
}

func InsertUser(ctx context.Context, db *pgxpool.Pool, mobile, passwordHash, nickname, avatar string) (uint64, error) {
	var id uint64
	err := db.QueryRow(
		ctx,
		`INSERT INTO "user" (mobile, password_hash, nickname, avatar) VALUES ($1, $2, $3, $4) RETURNING id`,
		mobile, passwordHash, nickname, avatar,
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func GetLoginInfo(ctx context.Context, db *pgxpool.Pool, mobile string) (uint64, string, string, string, error) {
	var (
		id           uint64
		passwordHash string
		nickname     string
		avatar       string
	)
	err := db.QueryRow(
		ctx,
		`SELECT id, password_hash, nickname, avatar FROM "user" WHERE mobile=$1`,
		mobile,
	).Scan(&id, &passwordHash, &nickname, &avatar)
	if err != nil {
		return 0, "", "", "", err
	}
	return id, passwordHash, nickname, avatar, nil
}

func GetUserInfo(ctx context.Context, db *pgxpool.Pool, userID uint64) (*user.UserInfo, error) {
	var (
		mobile   string
		nickname string
		avatar   string
		audit    int32
	)
	err := db.QueryRow(
		ctx,
		`SELECT mobile, nickname, avatar, audit_status FROM "user" WHERE id=$1`,
		userID,
	).Scan(&mobile, &nickname, &avatar, &audit)
	if err != nil {
		return nil, err
	}

	return &user.UserInfo{
		Id:          userID,
		Mobile:      mobile,
		Nickname:    nickname,
		Avatar:      avatar,
		AuditStatus: audit,
	}, nil
}

func UpdateUserInfo(ctx context.Context, db *pgxpool.Pool, userID uint64, nickname, avatar string) (bool, error) {
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Update(`"user"`).
		Set("audit_status", 0).
		Set("updated_at", sq.Expr("now()")).
		Where(sq.Eq{"id": userID})

	if nickname != "" {
		builder = builder.Set("nickname", nickname)
	}
	if avatar != "" {
		builder = builder.Set("avatar", avatar)
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return false, err
	}
	cmd, err := db.Exec(ctx, query, args...)
	if err != nil {
		return false, err
	}
	return cmd.RowsAffected() > 0, nil
}

func UpdateAuditStatus(ctx context.Context, db *pgxpool.Pool, userID uint64, auditStatus int32) (bool, error) {
	cmd, err := db.Exec(
		ctx,
		`UPDATE "user" SET audit_status=$1, updated_at=now() WHERE id=$2`,
		auditStatus, userID,
	)
	if err != nil {
		return false, err
	}
	return cmd.RowsAffected() > 0, nil
}
