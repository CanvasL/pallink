package dao

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type queryer interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type execer interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

func GetActivityEnrollInfo(ctx context.Context, q queryer, activityID uint64) (int32, int32, int32, error) {
	var maxPeople, currentPeople, status int32
	err := q.QueryRow(
		ctx,
		`SELECT max_people, current_people, status FROM activity WHERE id=$1`,
		activityID,
	).Scan(&maxPeople, &currentPeople, &status)
	if err != nil {
		return 0, 0, 0, err
	}
	return maxPeople, currentPeople, status, nil
}

func GetEnrollmentStatus(ctx context.Context, q queryer, activityID, userID uint64) (int32, bool, error) {
	var status int32
	err := q.QueryRow(
		ctx,
		`SELECT status FROM enrollment WHERE activity_id=$1 AND user_id=$2`,
		activityID, userID,
	).Scan(&status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, false, nil
		}
		return 0, false, err
	}
	return status, true, nil
}

func UpdateEnrollmentStatus(ctx context.Context, e execer, activityID, userID uint64, status int32, enrollTime *time.Time, checkinTime *time.Time) error {
	_, err := e.Exec(
		ctx,
		`UPDATE enrollment SET status=$1, enroll_time=COALESCE($2, enroll_time), checkin_time=$3 WHERE activity_id=$4 AND user_id=$5`,
		status, enrollTime, checkinTime, activityID, userID,
	)
	return err
}

func InsertEnrollment(ctx context.Context, e execer, activityID, userID uint64, status int32, enrollTime time.Time) error {
	_, err := e.Exec(
		ctx,
		`INSERT INTO enrollment (activity_id, user_id, status, enroll_time) VALUES ($1, $2, $3, $4)`,
		activityID, userID, status, enrollTime,
	)
	return err
}

func UpdateActivityPeople(ctx context.Context, e execer, activityID uint64, delta int32) error {
	if delta == 0 {
		return nil
	}
	if delta > 0 {
		_, err := e.Exec(ctx, `UPDATE activity SET current_people = current_people + $1 WHERE id=$2`, delta, activityID)
		return err
	}
	_, err := e.Exec(ctx, `UPDATE activity SET current_people = GREATEST(current_people + $1, 0) WHERE id=$2`, delta, activityID)
	return err
}
