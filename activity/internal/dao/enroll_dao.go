package dao

import (
	"context"
	"errors"
	"time"

	"pallink/activity/internal/dao/sqlc"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func GetActivityEnrollInfo(ctx context.Context, db sqlc.DBTX, activityID uint64) (int32, int32, int32, error) {
	q := sqlc.New(db)
	row, err := q.GetActivityEnrollInfo(ctx, int64(activityID))
	if err != nil {
		return 0, 0, 0, err
	}
	return row.MaxPeople, row.CurrentPeople, int32(row.Status), nil
}

func GetActivityCheckInInfo(ctx context.Context, db sqlc.DBTX, activityID uint64) (time.Time, int32, error) {
	q := sqlc.New(db)
	row, err := q.GetActivityCheckInInfo(ctx, int64(activityID))
	if err != nil {
		return time.Time{}, 0, err
	}
	if !row.StartTime.Valid {
		return time.Time{}, 0, errors.New("activity start_time required")
	}
	return row.StartTime.Time, int32(row.Status), nil
}

func GetEnrollmentStatus(ctx context.Context, db sqlc.DBTX, activityID, userID uint64) (int32, bool, error) {
	q := sqlc.New(db)
	status, err := q.GetEnrollmentStatus(ctx, sqlc.GetEnrollmentStatusParams{
		ActivityID: int64(activityID),
		UserID:     int64(userID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, false, nil
		}
		return 0, false, err
	}
	return int32(status), true, nil
}

func UpdateEnrollmentStatus(ctx context.Context, db sqlc.DBTX, activityID, userID uint64, status int32, enrollTime *time.Time, checkinTime *time.Time) error {
	q := sqlc.New(db)
	switch {
	case enrollTime != nil:
		return q.UpdateEnrollmentStatusWithEnrollTime(ctx, sqlc.UpdateEnrollmentStatusWithEnrollTimeParams{
			ActivityID: int64(activityID),
			UserID:     int64(userID),
			Status:     int16(status),
			EnrollTime: pgtype.Timestamptz{Time: *enrollTime, Valid: true},
		})
	case checkinTime != nil:
		return q.UpdateEnrollmentStatusWithCheckinTime(ctx, sqlc.UpdateEnrollmentStatusWithCheckinTimeParams{
			ActivityID:  int64(activityID),
			UserID:      int64(userID),
			Status:      int16(status),
			CheckinTime: pgtype.Timestamptz{Time: *checkinTime, Valid: true},
		})
	default:
		return q.UpdateEnrollmentStatus(ctx, sqlc.UpdateEnrollmentStatusParams{
			ActivityID: int64(activityID),
			UserID:     int64(userID),
			Status:     int16(status),
		})
	}
}

func InsertEnrollment(ctx context.Context, db sqlc.DBTX, activityID, userID uint64, status int32, enrollTime time.Time) error {
	q := sqlc.New(db)
	return q.InsertEnrollment(ctx, sqlc.InsertEnrollmentParams{
		ActivityID: int64(activityID),
		UserID:     int64(userID),
		Status:     int16(status),
		EnrollTime: pgtype.Timestamptz{Time: enrollTime, Valid: true},
	})
}

func UpdateActivityPeople(ctx context.Context, db sqlc.DBTX, activityID uint64, delta int32) error {
	if delta == 0 {
		return nil
	}
	q := sqlc.New(db)
	return q.UpdateActivityPeople(ctx, sqlc.UpdateActivityPeopleParams{
		ActivityID: int64(activityID),
		Delta:      delta,
	})
}
