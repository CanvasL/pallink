CREATE TABLE IF NOT EXISTS "user" (
    id BIGSERIAL PRIMARY KEY,
    mobile VARCHAR(20) NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    nickname VARCHAR(50) NOT NULL,
    avatar VARCHAR(255) NOT NULL DEFAULT '',
    audit_status SMALLINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS activity (
    id BIGSERIAL PRIMARY KEY,
    creator_id BIGINT NOT NULL REFERENCES "user"(id),
    title VARCHAR(100) NOT NULL,
    description TEXT,
    location VARCHAR(255) NOT NULL,
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ NOT NULL,
    max_people INTEGER NOT NULL DEFAULT 0,
    current_people INTEGER NOT NULL DEFAULT 0,
    status SMALLINT NOT NULL DEFAULT 1,
    audit_status SMALLINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS enrollment (
    id BIGSERIAL PRIMARY KEY,
    activity_id BIGINT NOT NULL REFERENCES activity(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    status SMALLINT NOT NULL DEFAULT 1,
    enroll_time TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    checkin_time TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT uk_activity_user UNIQUE (activity_id, user_id)
);

CREATE TABLE IF NOT EXISTS activity_comment (
    id BIGSERIAL PRIMARY KEY,
    activity_id BIGINT NOT NULL REFERENCES activity(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    parent_id BIGINT REFERENCES activity_comment(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    audit_status SMALLINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_user_mobile ON "user" (mobile);
CREATE INDEX IF NOT EXISTS idx_user_audit_status ON "user" (audit_status);
CREATE INDEX IF NOT EXISTS idx_activity_creator ON activity (creator_id);
CREATE INDEX IF NOT EXISTS idx_activity_status ON activity (status);
CREATE INDEX IF NOT EXISTS idx_activity_audit_status ON activity (audit_status);
CREATE INDEX IF NOT EXISTS idx_enrollment_user ON enrollment (user_id);
CREATE INDEX IF NOT EXISTS idx_comment_activity ON activity_comment (activity_id);
CREATE INDEX IF NOT EXISTS idx_comment_parent ON activity_comment (parent_id);
CREATE INDEX IF NOT EXISTS idx_comment_audit_status ON activity_comment (audit_status);
