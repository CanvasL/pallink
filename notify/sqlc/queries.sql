-- name: CountNotifications :one
SELECT count(*)
FROM notification
WHERE user_id = @user_id
  AND (NOT @unread_only OR read_at IS NULL);

-- name: ListNotifications :many
SELECT id, user_id, actor_id, type,
       COALESCE(activity_id, 0) AS activity_id,
       COALESCE(comment_id, 0) AS comment_id,
       COALESCE(parent_id, 0) AS parent_id,
       content, created_at, read_at
FROM notification
WHERE user_id = @user_id
  AND (NOT @unread_only OR read_at IS NULL)
ORDER BY created_at DESC
LIMIT @page_limit OFFSET @page_offset;

-- name: MarkReadAll :exec
UPDATE notification
SET read_at = now()
WHERE user_id = @user_id
  AND read_at IS NULL;

-- name: MarkReadOne :exec
UPDATE notification
SET read_at = now()
WHERE user_id = @user_id
  AND id = @id;

-- name: InsertNotification :exec
INSERT INTO notification (user_id, actor_id, type, activity_id, comment_id, parent_id, content, created_at)
VALUES (@user_id, @actor_id, @type, @activity_id, @comment_id, @parent_id, @content, now());
