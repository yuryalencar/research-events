-- +goose Up
-- +goose StatementBegin
CREATE TABLE deadlines (
    id                BIGSERIAL    PRIMARY KEY,
    event_id          BIGINT       NOT NULL REFERENCES events(id),
    type              VARCHAR(50)  NOT NULL,
    description       VARCHAR(255) NOT NULL,
    date              TIMESTAMPTZ  NOT NULL,
    is_optional       BOOLEAN      NOT NULL DEFAULT FALSE,

    -- false when superseded by a newer deadline of the same type. The UI shows
    -- only is_active=true deadlines by default, with a "view history" toggle.
    is_active         BOOLEAN      NOT NULL DEFAULT TRUE,
    superseded_by_id  BIGINT       REFERENCES deadlines(id),

    created_by_id     BIGINT       NOT NULL REFERENCES users(id),

    -- GORM-compatible timestamps (soft delete via deleted_at).
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at        TIMESTAMPTZ,

    CONSTRAINT deadlines_type_check CHECK (type IN (
        'abstract', 'paper', 'notification', 'camera_ready', 'other'
    ))
);

-- Fetching all deadlines for an event (grouped by type, active + history).
CREATE INDEX deadlines_event_id_idx ON deadlines (event_id);

-- "Show only active deadlines by default" is the common-path query.
CREATE INDEX deadlines_event_active_idx ON deadlines (event_id, is_active);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS deadlines;
-- +goose StatementEnd
