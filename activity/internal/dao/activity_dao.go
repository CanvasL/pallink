package dao

import (
	"context"
	"errors"
	"time"

	"pallink/activity/activity"
	"pallink/activity/internal/dao/sqlc"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
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

func CreateActivity(ctx context.Context, db sqlc.DBTX, creatorID uint64, title, description, location string, startTime, endTime time.Time, maxPeople int32) (uint64, error) {
	q := sqlc.New(db)
	id, err := q.CreateActivity(ctx, sqlc.CreateActivityParams{
		CreatorID:   int64(creatorID),
		Title:       title,
		Description: pgtype.Text{String: description, Valid: true},
		Location:    location,
		StartTime:   pgtype.Timestamptz{Time: startTime, Valid: true},
		EndTime:     pgtype.Timestamptz{Time: endTime, Valid: true},
		MaxPeople:   maxPeople,
	})
	if err != nil {
		return 0, err
	}
	return uint64(id), nil
}

func QueryActivityList(ctx context.Context, db sqlc.DBTX, filter ListFilter, viewerUserID uint64, page, pageSize int32) ([]*activity.ActivityInfo, int32, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	status := int32(-1)
	if filter.Status != nil {
		status = *filter.Status
	}
	auditStatus := int32(-1)
	if filter.AuditStatus != nil {
		auditStatus = *filter.AuditStatus
	}

	q := sqlc.New(db)
	total64, err := q.CountActivities(ctx, sqlc.CountActivitiesParams{
		EnrolledUserID: int64(filter.EnrolledUserID),
		UseCreator:     filter.UseCreatorID,
		CreatorID:      int64(filter.CreatorID),
		UseEnrolled:    filter.UseEnrolledUser,
		Status:         status,
		AuditStatus:    auditStatus,
		Keyword:        filter.Keyword,
	})
	if err != nil {
		return nil, 0, err
	}

	rows, err := q.ListActivities(ctx, sqlc.ListActivitiesParams{
		EnrolledUserID: int64(filter.EnrolledUserID),
		ViewerUserID:   int64(viewerUserID),
		UseCreator:     filter.UseCreatorID,
		CreatorID:      int64(filter.CreatorID),
		UseEnrolled:    filter.UseEnrolledUser,
		Status:         status,
		AuditStatus:    auditStatus,
		Keyword:        filter.Keyword,
		PageOffset:     (page - 1) * pageSize,
		PageLimit:      pageSize,
	})
	if err != nil {
		return nil, 0, err
	}

	list := make([]*activity.ActivityInfo, 0, len(rows))
	for _, row := range rows {
		description := ""
		if row.Description.Valid {
			description = row.Description.String
		}
		item := activity.ActivityInfo{
			Id:            uint64(row.ID),
			CreatorId:     uint64(row.CreatorID),
			Title:         row.Title,
			Description:   description,
			Location:      row.Location,
			MaxPeople:     row.MaxPeople,
			CurrentPeople: row.CurrentPeople,
			Status:        int32(row.Status),
			AuditStatus:   int32(row.AuditStatus),
			IsEnrolled:    row.IsEnrolled,
			StartTime:     timestamppb.New(timeFromTimestamptz(row.StartTime)),
			EndTime:       timestamppb.New(timeFromTimestamptz(row.EndTime)),
			CreatedAt:     timestamppb.New(timeFromTimestamptz(row.CreatedAt)),
		}
		list = append(list, &item)
	}

	return list, int32(total64), nil
}

func QueryActivityDetail(ctx context.Context, db sqlc.DBTX, activityID uint64, viewerUserID uint64) (*activity.ActivityInfo, error) {
	if activityID == 0 {
		return nil, errors.New("id required")
	}
	q := sqlc.New(db)
	row, err := q.GetActivityDetail(ctx, sqlc.GetActivityDetailParams{
		ID:           int64(activityID),
		ViewerUserID: int64(viewerUserID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("activity not found")
		}
		return nil, err
	}

	description := ""
	if row.Description.Valid {
		description = row.Description.String
	}
	info := activity.ActivityInfo{
		Id:            uint64(row.ID),
		CreatorId:     uint64(row.CreatorID),
		Title:         row.Title,
		Description:   description,
		Location:      row.Location,
		MaxPeople:     row.MaxPeople,
		CurrentPeople: row.CurrentPeople,
		Status:        int32(row.Status),
		AuditStatus:   int32(row.AuditStatus),
		IsEnrolled:    boolFromInterface(row.IsEnrolled),
		StartTime:     timestamppb.New(timeFromTimestamptz(row.StartTime)),
		EndTime:       timestamppb.New(timeFromTimestamptz(row.EndTime)),
		CreatedAt:     timestamppb.New(timeFromTimestamptz(row.CreatedAt)),
	}

	return &info, nil
}

func QueryParticipants(ctx context.Context, db sqlc.DBTX, activityID uint64, page, pageSize int32) ([]*activity.ParticipantInfo, int32, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	q := sqlc.New(db)
	total64, err := q.CountParticipants(ctx, int64(activityID))
	if err != nil {
		return nil, 0, err
	}

	rows, err := q.ListParticipants(ctx, sqlc.ListParticipantsParams{
		ActivityID: int64(activityID),
		PageOffset: (page - 1) * pageSize,
		PageLimit:  pageSize,
	})
	if err != nil {
		return nil, 0, err
	}

	list := make([]*activity.ParticipantInfo, 0, len(rows))
	for _, row := range rows {
		item := activity.ParticipantInfo{
			UserId: uint64(row.UserID),
			Status: int32(row.Status),
		}
		item.EnrollTime = timestamppb.New(timeFromTimestamptz(row.EnrollTime))
		if row.CheckinTime.Valid {
			item.CheckinTime = timestamppb.New(timeFromTimestamptz(row.CheckinTime))
		}
		list = append(list, &item)
	}

	return list, int32(total64), nil
}

func UpdateActivity(ctx context.Context, db sqlc.DBTX, in *activity.UpdateActivityRequest) (bool, error) {
	q := sqlc.New(db)
	affected, err := q.UpdateActivity(ctx, sqlc.UpdateActivityParams{
		ID:          int64(in.Id),
		CreatorID:   int64(in.CreatorId),
		Title:       optionalString(in.Title),
		Description: optionalString(in.Description),
		Location:    optionalString(in.Location),
		StartTime:   optionalTime(in.StartTime),
		EndTime:     optionalTime(in.EndTime),
		MaxPeople:   optionalInt32(in.MaxPeople),
		Status:      optionalInt32ToInt16(in.Status),
	})
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}

func UpdateAuditStatus(ctx context.Context, db sqlc.DBTX, activityID uint64, auditStatus int32) (bool, error) {
	q := sqlc.New(db)
	affected, err := q.UpdateActivityAuditStatus(ctx, sqlc.UpdateActivityAuditStatusParams{
		ID:          int64(activityID),
		AuditStatus: int16(auditStatus),
	})
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}

func GetActivityCreator(ctx context.Context, db sqlc.DBTX, activityID uint64) (uint64, error) {
	q := sqlc.New(db)
	creatorID, err := q.GetActivityCreator(ctx, int64(activityID))
	if err != nil {
		return 0, err
	}
	return uint64(creatorID), nil
}

func optionalString(val string) any {
	if val == "" {
		return nil
	}
	return val
}

func optionalTime(ts *timestamppb.Timestamp) pgtype.Timestamptz {
	if ts == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: ts.AsTime(), Valid: true}
}

func optionalInt32(val int32) any {
	if val == -1 {
		return nil
	}
	return val
}

func optionalInt32ToInt16(val int32) any {
	if val == -1 {
		return nil
	}
	return int16(val)
}

func timeFromTimestamptz(ts pgtype.Timestamptz) time.Time {
	if !ts.Valid {
		return time.Time{}
	}
	return ts.Time
}

func boolFromInterface(val any) bool {
	switch v := val.(type) {
	case bool:
		return v
	case int16:
		return v != 0
	case int32:
		return v != 0
	case int64:
		return v != 0
	default:
		return false
	}
}
