package logic

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"pallink/activity/activity"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func createComment(ctx context.Context, db *pgxpool.Pool, activityID, userID, parentID uint64, content string) (uint64, error) {
	if activityID == 0 || userID == 0 {
		return 0, errors.New("activity_id/user_id required")
	}
	if strings.TrimSpace(content) == "" {
		return 0, errors.New("content required")
	}

	if parentID > 0 {
		var parentActivityID uint64
		err := db.QueryRow(ctx, `SELECT activity_id FROM activity_comment WHERE id=$1`, parentID).Scan(&parentActivityID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return 0, errors.New("parent comment not found")
			}
			return 0, err
		}
		if parentActivityID != activityID {
			return 0, errors.New("parent comment not in activity")
		}
	}

	var id uint64
	var parent any
	if parentID > 0 {
		parent = parentID
	} else {
		parent = nil
	}
	err := db.QueryRow(
		ctx,
		`INSERT INTO activity_comment (activity_id, user_id, parent_id, content, audit_status, created_at)
		 VALUES ($1,$2,$3,$4,0,now()) RETURNING id`,
		activityID, userID, parent, content,
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func queryComments(ctx context.Context, db *pgxpool.Pool, activityID, parentID uint64, viewerUserID uint64, page, pageSize int32) ([]*activity.CommentInfo, int32, error) {
	if activityID == 0 {
		return nil, 0, errors.New("activity_id required")
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	where := []string{"activity_id=$1"}
	args := []any{activityID}
	if parentID > 0 {
		where = append(where, "parent_id=$2")
		args = append(args, parentID)
	} else {
		where = append(where, "parent_id IS NULL")
	}
	if viewerUserID == 0 {
		where = append(where, "audit_status=1")
	}

	countSQL := "SELECT count(*) FROM activity_comment WHERE " + strings.Join(where, " AND ")
	var total int32
	if err := db.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	limitArg := fmt.Sprintf("$%d", len(args)+1)
	offsetArg := fmt.Sprintf("$%d", len(args)+2)
	listSQL := "SELECT id, activity_id, user_id, COALESCE(parent_id,0), content, audit_status, created_at FROM activity_comment WHERE " +
		strings.Join(where, " AND ") + " ORDER BY created_at DESC LIMIT " + limitArg + " OFFSET " + offsetArg
	args = append(args, pageSize, (page-1)*pageSize)

	rows, err := db.Query(ctx, listSQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	list := make([]*activity.CommentInfo, 0)
	for rows.Next() {
		var (
			item      activity.CommentInfo
			createdAt time.Time
		)
		if err := rows.Scan(
			&item.Id,
			&item.ActivityId,
			&item.UserId,
			&item.ParentId,
			&item.Content,
			&item.AuditStatus,
			&createdAt,
		); err != nil {
			return nil, 0, err
		}
		item.CreatedAt = timestamppb.New(createdAt)
		list = append(list, &item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func updateCommentAuditStatus(ctx context.Context, db *pgxpool.Pool, commentID uint64, auditStatus int32) (bool, error) {
	if commentID == 0 {
		return false, errors.New("comment_id required")
	}
	if auditStatus < 0 || auditStatus > 2 {
		return false, errors.New("invalid audit_status")
	}

	cmd, err := db.Exec(
		ctx,
		`UPDATE activity_comment SET audit_status=$1 WHERE id=$2`,
		auditStatus, commentID,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return cmd.RowsAffected() > 0, nil
}
