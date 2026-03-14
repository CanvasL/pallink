package dao

import (
	"context"
	"errors"
	"strings"

	"pallink/activity/activity"
	"pallink/activity/internal/dao/sqlc"

	"github.com/jackc/pgx/v5"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func CreateComment(ctx context.Context, db sqlc.DBTX, activityID, userID, parentID uint64, content string) (uint64, uint64, error) {
	if activityID == 0 || userID == 0 {
		return 0, 0, errors.New("activity_id/user_id required")
	}
	if strings.TrimSpace(content) == "" {
		return 0, 0, errors.New("content required")
	}

	q := sqlc.New(db)
	var parentUserID uint64
	if parentID > 0 {
		parent, err := q.GetCommentParent(ctx, int64(parentID))
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return 0, 0, errors.New("parent comment not found")
			}
			return 0, 0, err
		}
		if uint64(parent.ActivityID) != activityID {
			return 0, 0, errors.New("parent comment not in activity")
		}
		parentUserID = uint64(parent.UserID)
	}

	id, err := q.InsertComment(ctx, sqlc.InsertCommentParams{
		ActivityID: int64(activityID),
		UserID:     int64(userID),
		ParentID:   int64(parentID),
		Content:    content,
	})
	if err != nil {
		return 0, 0, err
	}
	return uint64(id), parentUserID, nil
}

func QueryComments(ctx context.Context, db sqlc.DBTX, activityID, parentID uint64, viewerUserID uint64, page, pageSize int32) ([]*activity.CommentInfo, int32, error) {
	if activityID == 0 {
		return nil, 0, errors.New("activity_id required")
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	q := sqlc.New(db)
	includeUnapproved := viewerUserID > 0

	total64, err := q.CountComments(ctx, sqlc.CountCommentsParams{
		ActivityID:        int64(activityID),
		ParentID:          int64(parentID),
		IncludeUnapproved: includeUnapproved,
	})
	if err != nil {
		return nil, 0, err
	}

	rows, err := q.ListComments(ctx, sqlc.ListCommentsParams{
		ActivityID:        int64(activityID),
		ParentID:          int64(parentID),
		IncludeUnapproved: includeUnapproved,
		PageOffset:        (page - 1) * pageSize,
		PageLimit:         pageSize,
	})
	if err != nil {
		return nil, 0, err
	}

	list := make([]*activity.CommentInfo, 0, len(rows))
	for _, row := range rows {
		item := activity.CommentInfo{
			Id:          uint64(row.ID),
			ActivityId:  uint64(row.ActivityID),
			UserId:      uint64(row.UserID),
			ParentId:    uint64(row.ParentID),
			Content:     row.Content,
			AuditStatus: int32(row.AuditStatus),
			CreatedAt:   timestamppb.New(timeFromTimestamptz(row.CreatedAt)),
		}
		list = append(list, &item)
	}

	return list, int32(total64), nil
}

func UpdateCommentAuditStatus(ctx context.Context, db sqlc.DBTX, commentID uint64, auditStatus int32) (bool, error) {
	if commentID == 0 {
		return false, errors.New("comment_id required")
	}
	if auditStatus < 0 || auditStatus > 2 {
		return false, errors.New("invalid audit_status")
	}

	q := sqlc.New(db)
	affected, err := q.UpdateCommentAuditStatus(ctx, sqlc.UpdateCommentAuditStatusParams{
		ID:          int64(commentID),
		AuditStatus: int16(auditStatus),
	})
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}
