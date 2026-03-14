-- name: CreateActivity :one
INSERT INTO activity (creator_id, title, description, location, start_time, end_time, max_people, current_people, status, audit_status, created_at)
VALUES (@creator_id, @title, @description, @location, @start_time, @end_time, @max_people, 0, 1, 0, now())
RETURNING id;

-- name: CountActivities :one
SELECT count(*)
FROM activity a
LEFT JOIN enrollment e_filter ON e_filter.activity_id = a.id AND e_filter.user_id = @enrolled_user_id AND e_filter.status IN (1,2)
WHERE (NOT @use_creator OR a.creator_id = @creator_id)
  AND (NOT @use_enrolled OR e_filter.id IS NOT NULL)
  AND (@status = -1 OR a.status = @status)
  AND (@audit_status = -1 OR a.audit_status = @audit_status)
  AND (@keyword = '' OR a.title ILIKE '%' || @keyword || '%' OR a.description ILIKE '%' || @keyword || '%');

-- name: ListActivities :many
SELECT a.id,
       a.creator_id,
       a.title,
       a.description,
       a.location,
       a.start_time,
       a.end_time,
       a.max_people,
       a.current_people,
       a.status,
       a.audit_status,
       a.created_at,
       CASE WHEN e_viewer.id IS NULL THEN false ELSE true END AS is_enrolled
FROM activity a
LEFT JOIN enrollment e_filter ON e_filter.activity_id = a.id AND e_filter.user_id = @enrolled_user_id AND e_filter.status IN (1,2)
LEFT JOIN enrollment e_viewer ON e_viewer.activity_id = a.id AND e_viewer.user_id = @viewer_user_id AND e_viewer.status IN (1,2)
WHERE (NOT @use_creator OR a.creator_id = @creator_id)
  AND (NOT @use_enrolled OR e_filter.id IS NOT NULL)
  AND (@status = -1 OR a.status = @status)
  AND (@audit_status = -1 OR a.audit_status = @audit_status)
  AND (@keyword = '' OR a.title ILIKE '%' || @keyword || '%' OR a.description ILIKE '%' || @keyword || '%')
ORDER BY a.start_time DESC
LIMIT @page_limit OFFSET @page_offset;

-- name: GetActivityDetail :one
SELECT a.id,
       a.creator_id,
       a.title,
       a.description,
       a.location,
       a.start_time,
       a.end_time,
       a.max_people,
       a.current_people,
       a.status,
       a.audit_status,
       a.created_at,
       CASE WHEN @viewer_user_id = 0 THEN false
            ELSE EXISTS(
                SELECT 1 FROM enrollment e
                WHERE e.activity_id = a.id
                  AND e.user_id = @viewer_user_id
                  AND e.status IN (1,2)
            )
       END AS is_enrolled
FROM activity a
WHERE a.id = @id;

-- name: CountParticipants :one
SELECT count(*)
FROM enrollment
WHERE activity_id = @activity_id
  AND status IN (1,2);

-- name: ListParticipants :many
SELECT user_id, enroll_time, checkin_time, status
FROM enrollment
WHERE activity_id = @activity_id
  AND status IN (1,2)
ORDER BY enroll_time DESC
LIMIT @page_limit OFFSET @page_offset;

-- name: UpdateActivity :execrows
UPDATE activity
SET title = COALESCE(NULLIF(@title, ''), title),
    description = COALESCE(NULLIF(@description, ''), description),
    location = COALESCE(NULLIF(@location, ''), location),
    start_time = COALESCE(@start_time, start_time),
    end_time = COALESCE(@end_time, end_time),
    max_people = COALESCE(NULLIF(@max_people, -1), max_people),
    status = COALESCE(NULLIF(@status, -1), status),
    audit_status = 0,
    updated_at = now()
WHERE id = @id AND creator_id = @creator_id;

-- name: UpdateActivityAuditStatus :execrows
UPDATE activity
SET audit_status = @audit_status,
    updated_at = now()
WHERE id = @id;

-- name: GetActivityCreator :one
SELECT creator_id FROM activity WHERE id = @id;

-- name: GetActivityEnrollInfo :one
SELECT max_people, current_people, status
FROM activity
WHERE id = @activity_id;

-- name: GetActivityEnrollInfoForUpdate :one
SELECT max_people, current_people, status
FROM activity
WHERE id = @activity_id
FOR UPDATE;

-- name: GetActivityCheckInInfo :one
SELECT start_time, status
FROM activity
WHERE id = @activity_id;

-- name: GetEnrollmentStatus :one
SELECT status
FROM enrollment
WHERE activity_id = @activity_id
  AND user_id = @user_id;

-- name: GetEnrollmentStatusForUpdate :one
SELECT status
FROM enrollment
WHERE activity_id = @activity_id
  AND user_id = @user_id
FOR UPDATE;

-- name: UpdateEnrollmentStatus :exec
UPDATE enrollment
SET status = @status
WHERE activity_id = @activity_id
  AND user_id = @user_id;

-- name: UpdateEnrollmentStatusWithEnrollTime :exec
UPDATE enrollment
SET status = @status,
    enroll_time = @enroll_time
WHERE activity_id = @activity_id
  AND user_id = @user_id;

-- name: UpdateEnrollmentStatusWithCheckinTime :exec
UPDATE enrollment
SET status = @status,
    checkin_time = @checkin_time
WHERE activity_id = @activity_id
  AND user_id = @user_id;

-- name: InsertEnrollment :exec
INSERT INTO enrollment (activity_id, user_id, status, enroll_time)
VALUES (@activity_id, @user_id, @status, @enroll_time);

-- name: UpdateActivityPeople :exec
UPDATE activity
SET current_people = GREATEST(current_people + @delta, 0)
WHERE id = @activity_id;

-- name: GetCommentParent :one
SELECT activity_id, user_id
FROM activity_comment
WHERE id = @id;

-- name: InsertComment :one
INSERT INTO activity_comment (activity_id, user_id, parent_id, content, audit_status, created_at)
VALUES (@activity_id, @user_id, NULLIF(@parent_id, 0), @content, 0, now())
RETURNING id;

-- name: CountComments :one
SELECT count(*)
FROM activity_comment
WHERE activity_id = @activity_id
  AND (
    (@parent_id = 0 AND parent_id IS NULL)
    OR
    (@parent_id <> 0 AND parent_id = @parent_id)
  )
  AND (@include_unapproved OR audit_status = 1);

-- name: ListComments :many
SELECT id, activity_id, user_id, COALESCE(parent_id, 0) AS parent_id, content, audit_status, created_at
FROM activity_comment
WHERE activity_id = @activity_id
  AND (
    (@parent_id = 0 AND parent_id IS NULL)
    OR
    (@parent_id <> 0 AND parent_id = @parent_id)
  )
  AND (@include_unapproved OR audit_status = 1)
ORDER BY created_at DESC
LIMIT @page_limit OFFSET @page_offset;

-- name: UpdateCommentAuditStatus :execrows
UPDATE activity_comment
SET audit_status = @audit_status
WHERE id = @id;
