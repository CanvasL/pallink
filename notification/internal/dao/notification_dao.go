package dao

import (
	"context"
	"errors"
	"time"

	"pallink/notification/internal/dao/sqlc"
	"pallink/notification/notification"

	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ListFilter struct {
	UserID     uint64
	UnreadOnly bool
	Page       int32
	PageSize   int32
}

func QueryNotifications(ctx context.Context, db sqlc.DBTX, filter ListFilter) ([]*notification.NotificationInfo, int32, error) {
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

	q := sqlc.New(db)
	total64, err := q.CountNotifications(ctx, sqlc.CountNotificationsParams{
		UserID:     int64(filter.UserID),
		UnreadOnly: filter.UnreadOnly,
	})
	if err != nil {
		return nil, 0, err
	}

	rows, err := q.ListNotifications(ctx, sqlc.ListNotificationsParams{
		UserID:     int64(filter.UserID),
		UnreadOnly: filter.UnreadOnly,
		PageOffset: (page - 1) * pageSize,
		PageLimit:  pageSize,
	})
	if err != nil {
		return nil, 0, err
	}

	list := make([]*notification.NotificationInfo, 0, len(rows))
	for _, row := range rows {
		item := notification.NotificationInfo{
			Id:         uint64(row.ID),
			UserId:     uint64(row.UserID),
			ActorId:    uint64(row.ActorID),
			Type:       row.Type,
			ActivityId: uint64(row.ActivityID),
			CommentId:  uint64(row.CommentID),
			ParentId:   uint64(row.ParentID),
			Content:    row.Content,
			CreatedAt:  timestamppb.New(timeFromTimestamptz(row.CreatedAt)),
		}
		if row.ReadAt.Valid {
			item.ReadAt = timestamppb.New(timeFromTimestamptz(row.ReadAt))
		}
		list = append(list, &item)
	}

	return list, int32(total64), nil
}

func MarkRead(ctx context.Context, db sqlc.DBTX, userID, notificationID uint64) error {
	if userID == 0 {
		return errors.New("user_id required")
	}
	q := sqlc.New(db)
	if notificationID == 0 {
		return q.MarkReadAll(ctx, int64(userID))
	}
	return q.MarkReadOne(ctx, sqlc.MarkReadOneParams{
		UserID: int64(userID),
		ID:     int64(notificationID),
	})
}

func InsertNotification(ctx context.Context, db sqlc.DBTX, userID, actorID uint64, typ string, activityID, commentID, parentID uint64, content string) error {
	if userID == 0 || actorID == 0 {
		return errors.New("user_id/actor_id required")
	}
	if typ == "" {
		return errors.New("type required")
	}
	q := sqlc.New(db)
	return q.InsertNotification(ctx, sqlc.InsertNotificationParams{
		UserID:     int64(userID),
		ActorID:    int64(actorID),
		Type:       typ,
		ActivityID: optionalInt8(activityID),
		CommentID:  optionalInt8(commentID),
		ParentID:   optionalInt8(parentID),
		Content:    content,
	})
}

func optionalInt8(val uint64) pgtype.Int8 {
	if val == 0 {
		return pgtype.Int8{}
	}
	return pgtype.Int8{Int64: int64(val), Valid: true}
}

func timeFromTimestamptz(ts pgtype.Timestamptz) time.Time {
	if !ts.Valid {
		return time.Time{}
	}
	return ts.Time
}
