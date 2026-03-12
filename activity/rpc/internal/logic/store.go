package logic

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"pallink/activity/rpc/activity"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type listFilter struct {
	CreatorID       uint64
	UseCreatorID    bool
	EnrolledUserID  uint64
	UseEnrolledUser bool
	Status          *int32
	Keyword         string
}

func queryActivityList(ctx context.Context, db *pgxpool.Pool, filter listFilter, viewerUserID uint64, page, pageSize int32) ([]*activity.ActivityInfo, int32, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	args := make([]any, 0)
	addArg := func(v any) string {
		args = append(args, v)
		return fmt.Sprintf("$%d", len(args))
	}

	var sb strings.Builder
	sb.WriteString(" FROM activity a JOIN \"user\" u ON u.id=a.creator_id ")

	if filter.UseEnrolledUser {
		userArg := addArg(filter.EnrolledUserID)
		sb.WriteString(" JOIN enrollment e ON e.activity_id=a.id AND e.user_id=")
		sb.WriteString(userArg)
		sb.WriteString(" AND e.status IN (1,2) ")
	}

	var isEnrolledExpr string
	if viewerUserID > 0 {
		userArg := addArg(viewerUserID)
		sb.WriteString(" LEFT JOIN enrollment e2 ON e2.activity_id=a.id AND e2.user_id=")
		sb.WriteString(userArg)
		sb.WriteString(" AND e2.status IN (1,2) ")
		isEnrolledExpr = "(e2.id IS NOT NULL)"
	} else {
		isEnrolledExpr = "false"
	}

	where := []string{"1=1"}
	if filter.UseCreatorID {
		where = append(where, "a.creator_id="+addArg(filter.CreatorID))
	}
	if filter.Status != nil {
		where = append(where, "a.status="+addArg(*filter.Status))
	}
	if filter.Keyword != "" {
		kw := "%" + filter.Keyword + "%"
		where = append(where, "(a.title ILIKE "+addArg(kw)+" OR a.description ILIKE "+addArg(kw)+")")
	}

	countSQL := "SELECT count(*)" + sb.String() + " WHERE " + strings.Join(where, " AND ")
	var total int32
	if err := db.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	limitArg := addArg(pageSize)
	offsetArg := addArg((page - 1) * pageSize)
	listSQL := "SELECT a.id, a.creator_id, u.nickname, u.avatar, a.title, a.description, a.location, " +
		"a.start_time, a.end_time, a.max_people, a.current_people, a.status, a.created_at, " + isEnrolledExpr + " AS is_enrolled" +
		sb.String() + " WHERE " + strings.Join(where, " AND ") +
		" ORDER BY a.start_time DESC LIMIT " + limitArg + " OFFSET " + offsetArg

	rows, err := db.Query(ctx, listSQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	list := make([]*activity.ActivityInfo, 0)
	for rows.Next() {
		var (
			item          activity.ActivityInfo
			startTime     time.Time
			endTime       time.Time
			createdAt     time.Time
			creatorName   string
			creatorAvatar string
			isEnrolled    bool
		)
		if err := rows.Scan(
			&item.Id,
			&item.CreatorId,
			&creatorName,
			&creatorAvatar,
			&item.Title,
			&item.Description,
			&item.Location,
			&startTime,
			&endTime,
			&item.MaxPeople,
			&item.CurrentPeople,
			&item.Status,
			&createdAt,
			&isEnrolled,
		); err != nil {
			return nil, 0, err
		}
		item.StartTime = timestamppb.New(startTime)
		item.EndTime = timestamppb.New(endTime)
		item.CreatedAt = timestamppb.New(createdAt)
		item.CreatorName = creatorName
		item.CreatorAvatar = creatorAvatar
		item.IsEnrolled = isEnrolled
		list = append(list, &item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func queryActivityDetail(ctx context.Context, db *pgxpool.Pool, activityID uint64, viewerUserID uint64) (*activity.ActivityInfo, error) {
	if activityID == 0 {
		return nil, errors.New("id required")
	}

	var (
		info          activity.ActivityInfo
		startTime     time.Time
		endTime       time.Time
		createdAt     time.Time
		creatorName   string
		creatorAvatar string
	)
	if err := db.QueryRow(
		ctx,
		`SELECT a.id, a.creator_id, u.nickname, u.avatar, a.title, a.description, a.location,
		        a.start_time, a.end_time, a.max_people, a.current_people, a.status, a.created_at
		   FROM activity a
		   JOIN "user" u ON u.id=a.creator_id
		  WHERE a.id=$1`,
		activityID,
	).Scan(
		&info.Id,
		&info.CreatorId,
		&creatorName,
		&creatorAvatar,
		&info.Title,
		&info.Description,
		&info.Location,
		&startTime,
		&endTime,
		&info.MaxPeople,
		&info.CurrentPeople,
		&info.Status,
		&createdAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("activity not found")
		}
		return nil, err
	}

	info.StartTime = timestamppb.New(startTime)
	info.EndTime = timestamppb.New(endTime)
	info.CreatedAt = timestamppb.New(createdAt)
	info.CreatorName = creatorName
	info.CreatorAvatar = creatorAvatar

	if viewerUserID > 0 {
		var enrolled bool
		if err := db.QueryRow(
			ctx,
			`SELECT EXISTS(SELECT 1 FROM enrollment WHERE activity_id=$1 AND user_id=$2 AND status IN (1,2))`,
			activityID, viewerUserID,
		).Scan(&enrolled); err != nil {
			return nil, err
		}
		info.IsEnrolled = enrolled
	}

	return &info, nil
}

func queryParticipants(ctx context.Context, db *pgxpool.Pool, activityID uint64, page, pageSize int32) ([]*activity.ParticipantInfo, int32, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	var total int32
	if err := db.QueryRow(ctx, `SELECT count(*) FROM enrollment WHERE activity_id=$1 AND status IN (1,2)`, activityID).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := db.Query(
		ctx,
		`SELECT u.id, u.nickname, u.avatar, e.enroll_time, e.checkin_time, e.status
		   FROM enrollment e
		   JOIN "user" u ON u.id=e.user_id
		  WHERE e.activity_id=$1 AND e.status IN (1,2)
		  ORDER BY e.enroll_time DESC
		  LIMIT $2 OFFSET $3`,
		activityID, pageSize, (page-1)*pageSize,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	list := make([]*activity.ParticipantInfo, 0)
	for rows.Next() {
		var (
			item        activity.ParticipantInfo
			enrollTime  time.Time
			checkinTime *time.Time
		)
		if err := rows.Scan(&item.UserId, &item.Nickname, &item.Avatar, &enrollTime, &checkinTime, &item.Status); err != nil {
			return nil, 0, err
		}
		item.EnrollTime = timestamppb.New(enrollTime)
		if checkinTime != nil {
			item.CheckinTime = timestamppb.New(*checkinTime)
		}
		list = append(list, &item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
