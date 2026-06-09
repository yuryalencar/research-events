-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
    id                       BIGSERIAL    PRIMARY KEY,
    name                     VARCHAR(255) NOT NULL,
    email                    VARCHAR(255) NOT NULL,
    password_hash            VARCHAR(255),
    role                     VARCHAR(50)  NOT NULL DEFAULT 'contributor',

    -- Stateful JWT: one valid token pair per user at a time.
    -- access_token_jti stores the UUID jti claim of the current access token.
    -- Comparing this against the incoming JWT's jti detects revoked/logged-out tokens.
    access_token_jti         VARCHAR(36),
    access_token_expires_at  TIMESTAMPTZ,
    refresh_token_hash       VARCHAR(64),   -- SHA-256 hex of the current refresh token
    refresh_token_expires_at TIMESTAMPTZ,

    -- Account lockout: set after 5 consecutive failed login attempts.
    failed_login_attempts    INTEGER      NOT NULL DEFAULT 0,
    locked_at                TIMESTAMPTZ,

    -- GORM-compatible timestamps (soft delete via deleted_at).
    created_at               TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at               TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at               TIMESTAMPTZ,

    CONSTRAINT users_email_unique UNIQUE (email),
    CONSTRAINT users_role_check   CHECK  (role IN ('admin', 'moderator', 'contributor'))
);

-- Unique email index (enforces constraint + speeds up FindByEmail on every login).
CREATE UNIQUE INDEX users_email_idx ON users (email);

-- Role index for admin queries that filter by role (e.g. list all admins).
CREATE INDEX users_role_idx ON users (role);

-- JTI index: hit on every authenticated request to verify the token has not been revoked.
CREATE INDEX users_access_token_jti_idx ON users (access_token_jti);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
