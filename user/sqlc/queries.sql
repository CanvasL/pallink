-- name: GetUserIDByMobile :one
SELECT id FROM "user" WHERE mobile = @mobile;

-- name: InsertUser :one
INSERT INTO "user" (mobile, password_hash, nickname, avatar)
VALUES (@mobile, @password_hash, @nickname, @avatar)
RETURNING id;

-- name: GetLoginInfo :one
SELECT id, password_hash, nickname, avatar
FROM "user"
WHERE mobile = @mobile;

-- name: GetUserInfo :one
SELECT id, mobile, nickname, avatar, audit_status
FROM "user"
WHERE id = @id;

-- name: UpdateUserInfo :execrows
UPDATE "user"
SET nickname = COALESCE(NULLIF(@nickname, ''), nickname),
    avatar = COALESCE(NULLIF(@avatar, ''), avatar),
    audit_status = 0,
    updated_at = now()
WHERE id = @id;

-- name: UpdateUserAuditStatus :execrows
UPDATE "user"
SET audit_status = @audit_status,
    updated_at = now()
WHERE id = @id;
