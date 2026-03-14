-- name: UpsertConversation :one
INSERT INTO im_conversation (user1_id, user2_id, created_at)
VALUES (@user1_id, @user2_id, now())
ON CONFLICT (user1_id, user2_id)
DO UPDATE SET user1_id = EXCLUDED.user1_id
RETURNING id, user1_id, user2_id, created_at;

-- name: EnsureConversationMember :exec
INSERT INTO im_conversation_member (conversation_id, user_id, last_read_msg_id, updated_at)
VALUES (@conversation_id, @user_id, 0, now())
ON CONFLICT (conversation_id, user_id)
DO NOTHING;

-- name: GetConversationMember :one
SELECT conversation_id, user_id, last_read_msg_id, updated_at
FROM im_conversation_member
WHERE conversation_id = @conversation_id
  AND user_id = @user_id;

-- name: GetConversationSummary :one
SELECT c.id,
       (CASE WHEN c.user1_id = @user_id THEN c.user2_id ELSE c.user1_id END)::BIGINT AS peer_user_id,
       COALESCE(m.id, 0) AS last_message_id,
       COALESCE(m.sender_id, 0) AS last_sender_id,
       COALESCE(m.content, '') AS last_message,
       COALESCE(m.created_at, c.created_at) AS last_message_at,
       c.created_at,
       (
         SELECT count(*)
         FROM im_message imx
         WHERE imx.conversation_id = c.id
           AND imx.id > cm.last_read_msg_id
           AND imx.sender_id <> @user_id
       ) AS unread_count
FROM im_conversation_member cm
JOIN im_conversation c ON c.id = cm.conversation_id
LEFT JOIN LATERAL (
    SELECT id, sender_id, content, created_at
    FROM im_message
    WHERE conversation_id = c.id
    ORDER BY id DESC
    LIMIT 1
) m ON true
WHERE cm.user_id = @user_id
  AND c.id = @conversation_id;

-- name: CountConversations :one
SELECT count(*)
FROM im_conversation_member
WHERE user_id = @user_id;

-- name: ListConversations :many
SELECT c.id,
       (CASE WHEN c.user1_id = @user_id THEN c.user2_id ELSE c.user1_id END)::BIGINT AS peer_user_id,
       COALESCE(m.id, 0) AS last_message_id,
       COALESCE(m.sender_id, 0) AS last_sender_id,
       COALESCE(m.content, '') AS last_message,
       COALESCE(m.created_at, c.created_at) AS last_message_at,
       c.created_at,
       (
         SELECT count(*)
         FROM im_message imx
         WHERE imx.conversation_id = c.id
           AND imx.id > cm.last_read_msg_id
           AND imx.sender_id <> @user_id
       ) AS unread_count
FROM im_conversation_member cm
JOIN im_conversation c ON c.id = cm.conversation_id
LEFT JOIN LATERAL (
    SELECT id, sender_id, content, created_at
    FROM im_message
    WHERE conversation_id = c.id
    ORDER BY id DESC
    LIMIT 1
) m ON true
WHERE cm.user_id = @user_id
ORDER BY COALESCE(m.id, 0) DESC, c.id DESC
LIMIT @page_limit OFFSET @page_offset;

-- name: InsertMessage :one
INSERT INTO im_message (conversation_id, sender_id, content, audit_status, created_at)
VALUES (@conversation_id, @sender_id, @content, 1, now())
RETURNING id, conversation_id, sender_id, content, audit_status, created_at;

-- name: ListMessages :many
SELECT id, conversation_id, sender_id, content, audit_status, created_at
FROM im_message
WHERE conversation_id = @conversation_id
  AND (@before_message_id = 0 OR id < @before_message_id)
ORDER BY id DESC
LIMIT @page_limit;

-- name: GetLatestMessageID :one
SELECT id
FROM im_message
WHERE conversation_id = @conversation_id
ORDER BY id DESC
LIMIT 1;

-- name: GetMessageInConversation :one
SELECT id
FROM im_message
WHERE conversation_id = @conversation_id
  AND id = @message_id;

-- name: UpdateConversationReadCursor :execrows
UPDATE im_conversation_member
SET last_read_msg_id = GREATEST(last_read_msg_id, @last_read_msg_id),
    updated_at = now()
WHERE conversation_id = @conversation_id
  AND user_id = @user_id;
