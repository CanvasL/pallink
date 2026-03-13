package dao

import (
	"context"
	"errors"
	"time"

	"pallink/notify/notify"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ListFilter struct {
	UserID     uint64
	UnreadOnly bool
	Page       int32
	PageSize   int32
}

func QueryNotifications(ctx context.Context, db *pgxpool.Pool, filter ListFilter) ([]*notify.NotificationInfo, int32, error) {
	if filter.UserID == 0 {
		return nil, 0, errors.New("user_id required")
	}
	page := filter.Page
	pageSize := filter.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	base := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	countQ := base.Select("count(*)").From("notification").Where(sq.Eq{"user_id": filter.UserID})
	listQ := base.Select(
		"id",
		"user_id",
		"actor_id",
		"type",
		"COALESCE(activity_id,0)",
		"COALESCE(comment_id,0)",
		"COALESCE(parent_id,0)",
		"content",
		"created_at",
		"read_at",
	).From("notification").Where(sq.Eq{"user_id": filter.UserID})

	if filter.UnreadOnly {
		countQ = countQ.Where("read_at IS NULL")
		listQ = listQ.Where("read_at IS NULL")
	}

	countSQL, countArgs, err := countQ.ToSql()
	if err != nil {
		return nil, 0, err
	}
	var total int32
	if err := db.QueryRow(ctx, countSQL, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	listQ = listQ.OrderBy("created_at DESC").Limit(uint64(pageSize)).Offset(uint64((page - 1) * pageSize))
	listSQL, listArgs, err := listQ.ToSql()
	if err != nil {
		return nil, 0, err
	}

	rows, err := db.Query(ctx, listSQL, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	list := make([]*notify.NotificationInfo, 0)
	for rows.Next() {
		var (
			item      notify.NotificationInfo
			createdAt time.Time
			readAt    *time.Time
		)
		if err := rows.Scan(
			&item.Id,
			&item.UserId,
			&item.ActorId,
			&item.Type,
			&item.ActivityId,
			&item.CommentId,
			&item.ParentId,
			&item.Content,
			&createdAt,
			&readAt,
		); err != nil {
			return nil, 0, err
		}
		item.CreatedAt = timestamppb.New(createdAt)
		if readAt != nil {
			item.ReadAt = timestamppb.New(*readAt)
		}
		list = append(list, &item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func MarkRead(ctx context.Context, db *pgxpool.Pool, userID, notificationID uint64) error {
	if userID == 0 {
		return errors.New("user_id required")
	}
	if notificationID == 0 {
		_, err := db.Exec(ctx, `UPDATE notification SET read_at=now() WHERE user_id=$1 AND read_at IS NULL`, userID)
		return err
	}
	_, err := db.Exec(ctx, `UPDATE notification SET read_at=now() WHERE user_id=$1 AND id=$2`, userID, notificationID)
	return err
}

func InsertNotification(ctx context.Context, db *pgxpool.Pool, userID, actorID uint64, typ string, activityID, commentID, parentID uint64, content string) error {
	if userID == 0 || actorID == 0 {
		return errors.New("user_id/actor_id required")
	}
	if typ == "" {
		return errors.New("type required")
	}
	_, err := db.Exec(
		ctx,
		`INSERT INTO notification (user_id, actor_id, type, activity_id, comment_id, parent_id, content, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,now())`,
		userID, actorID, typ, activityID, commentID, parentID, content,
	)
	return err
}
