package dao

import (
	"context"
	"errors"
	"time"

	"pallink/activity/activity"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ListFilter struct {
	CreatorID       uint64
	UseCreatorID    bool
	EnrolledUserID  uint64
	UseEnrolledUser bool
	Status          *int32
	AuditStatus     *int32
	Keyword         string
}

func CreateActivity(ctx context.Context, db *pgxpool.Pool, creatorID uint64, title, description, location string, startTime, endTime time.Time, maxPeople int32) (uint64, error) {
	var id uint64
	err := db.QueryRow(
		ctx,
		`INSERT INTO activity (creator_id, title, description, location, start_time, end_time, max_people, current_people, status, audit_status, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,0,1,0,now())
		 RETURNING id`,
		creatorID, title, description, location, startTime, endTime, maxPeople,
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func QueryActivityList(ctx context.Context, db *pgxpool.Pool, filter ListFilter, viewerUserID uint64, page, pageSize int32) ([]*activity.ActivityInfo, int32, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	base := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	countQ := base.Select("count(*)").From("activity a")
	listQ := base.Select(
		"a.id",
		"a.creator_id",
		"a.title",
		"a.description",
		"a.location",
		"a.start_time",
		"a.end_time",
		"a.max_people",
		"a.current_people",
		"a.status",
		"a.audit_status",
		"a.created_at",
	).From("activity a")

	if filter.UseEnrolledUser {
		join := "enrollment e ON e.activity_id=a.id AND e.user_id=? AND e.status IN (1,2)"
		countQ = countQ.Join(join, filter.EnrolledUserID)
		listQ = listQ.Join(join, filter.EnrolledUserID)
	}

	if viewerUserID > 0 {
		join := "enrollment e2 ON e2.activity_id=a.id AND e2.user_id=? AND e2.status IN (1,2)"
		countQ = countQ.LeftJoin(join, viewerUserID)
		listQ = listQ.LeftJoin(join, viewerUserID).
			Column("CASE WHEN e2.id IS NULL THEN false ELSE true END AS is_enrolled")
	} else {
		listQ = listQ.Column("false AS is_enrolled")
	}

	if filter.UseCreatorID {
		countQ = countQ.Where(sq.Eq{"a.creator_id": filter.CreatorID})
		listQ = listQ.Where(sq.Eq{"a.creator_id": filter.CreatorID})
	}
	if filter.Status != nil {
		countQ = countQ.Where(sq.Eq{"a.status": *filter.Status})
		listQ = listQ.Where(sq.Eq{"a.status": *filter.Status})
	}
	if filter.AuditStatus != nil {
		countQ = countQ.Where(sq.Eq{"a.audit_status": *filter.AuditStatus})
		listQ = listQ.Where(sq.Eq{"a.audit_status": *filter.AuditStatus})
	}
	if filter.Keyword != "" {
		kw := "%" + filter.Keyword + "%"
		expr := sq.Expr("(a.title ILIKE ? OR a.description ILIKE ?)", kw, kw)
		countQ = countQ.Where(expr)
		listQ = listQ.Where(expr)
	}

	countSQL, countArgs, err := countQ.ToSql()
	if err != nil {
		return nil, 0, err
	}
	var total int32
	if err := db.QueryRow(ctx, countSQL, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	listQ = listQ.OrderBy("a.start_time DESC").Limit(uint64(pageSize)).Offset(uint64((page - 1) * pageSize))
	listSQL, listArgs, err := listQ.ToSql()
	if err != nil {
		return nil, 0, err
	}

	rows, err := db.Query(ctx, listSQL, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	list := make([]*activity.ActivityInfo, 0)
	for rows.Next() {
		var (
			item       activity.ActivityInfo
			startTime  time.Time
			endTime    time.Time
			createdAt  time.Time
			isEnrolled bool
		)
		if err := rows.Scan(
			&item.Id,
			&item.CreatorId,
			&item.Title,
			&item.Description,
			&item.Location,
			&startTime,
			&endTime,
			&item.MaxPeople,
			&item.CurrentPeople,
			&item.Status,
			&item.AuditStatus,
			&createdAt,
			&isEnrolled,
		); err != nil {
			return nil, 0, err
		}
		item.StartTime = timestamppb.New(startTime)
		item.EndTime = timestamppb.New(endTime)
		item.CreatedAt = timestamppb.New(createdAt)
		item.IsEnrolled = isEnrolled
		list = append(list, &item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func QueryActivityDetail(ctx context.Context, db *pgxpool.Pool, activityID uint64, viewerUserID uint64) (*activity.ActivityInfo, error) {
	if activityID == 0 {
		return nil, errors.New("id required")
	}

	var (
		info      activity.ActivityInfo
		startTime time.Time
		endTime   time.Time
		createdAt time.Time
	)
	if err := db.QueryRow(
		ctx,
		`SELECT a.id, a.creator_id, a.title, a.description, a.location,
	        a.start_time, a.end_time, a.max_people, a.current_people, a.status, a.audit_status, a.created_at
	   FROM activity a
	  WHERE a.id=$1`,
		activityID,
	).Scan(
		&info.Id,
		&info.CreatorId,
		&info.Title,
		&info.Description,
		&info.Location,
		&startTime,
		&endTime,
		&info.MaxPeople,
		&info.CurrentPeople,
		&info.Status,
		&info.AuditStatus,
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

func QueryParticipants(ctx context.Context, db *pgxpool.Pool, activityID uint64, page, pageSize int32) ([]*activity.ParticipantInfo, int32, error) {
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
		`SELECT e.user_id, e.enroll_time, e.checkin_time, e.status
		   FROM enrollment e
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
		if err := rows.Scan(&item.UserId, &enrollTime, &checkinTime, &item.Status); err != nil {
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

func UpdateActivity(ctx context.Context, db *pgxpool.Pool, in *activity.UpdateActivityRequest) (bool, error) {
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Update("activity").
		Set("updated_at", sq.Expr("now()")).
		Set("audit_status", 0).
		Where(sq.Eq{"id": in.Id, "creator_id": in.CreatorId})

	if in.Title != "" {
		builder = builder.Set("title", in.Title)
	}
	if in.Description != "" {
		builder = builder.Set("description", in.Description)
	}
	if in.Location != "" {
		builder = builder.Set("location", in.Location)
	}
	if in.StartTime != nil {
		builder = builder.Set("start_time", in.StartTime.AsTime())
	}
	if in.EndTime != nil {
		builder = builder.Set("end_time", in.EndTime.AsTime())
	}
	if in.MaxPeople != -1 {
		builder = builder.Set("max_people", in.MaxPeople)
	}
	if in.Status != -1 {
		builder = builder.Set("status", in.Status)
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

func UpdateAuditStatus(ctx context.Context, db *pgxpool.Pool, activityID uint64, auditStatus int32) (bool, error) {
	cmd, err := db.Exec(
		ctx,
		`UPDATE activity SET audit_status=$1, updated_at=now() WHERE id=$2`,
		auditStatus, activityID,
	)
	if err != nil {
		return false, err
	}
	return cmd.RowsAffected() > 0, nil
}

func GetActivityCreator(ctx context.Context, db *pgxpool.Pool, activityID uint64) (uint64, error) {
	var creatorID uint64
	if err := db.QueryRow(ctx, `SELECT creator_id FROM activity WHERE id=$1`, activityID).Scan(&creatorID); err != nil {
		return 0, err
	}
	return creatorID, nil
}
