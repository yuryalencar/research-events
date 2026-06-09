-- +goose Up
-- +goose StatementBegin
CREATE TABLE audit_logs (
    id             BIGSERIAL    PRIMARY KEY,
    entity_type    VARCHAR(50)  NOT NULL,
    entity_id      BIGINT       NOT NULL,
    action         VARCHAR(50)  NOT NULL,
    changed_by_id  BIGINT       NOT NULL REFERENCES users(id),
    diff           JSONB,

    -- GORM-compatible timestamps (soft delete via deleted_at).
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at     TIMESTAMPTZ,

    CONSTRAINT audit_logs_entity_type_check CHECK (entity_type IN ('event', 'deadline', 'user')),
    CONSTRAINT audit_logs_action_check      CHECK (action IN (
        'created', 'updated', 'approved', 'rejected',
        'deadline_added', 'deadline_superseded', 'unlocked'
    ))
);

-- Index for fetching the full audit trail of a specific entity (e.g. event detail page).
CREATE INDEX audit_logs_entity_idx ON audit_logs (entity_type, entity_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS audit_logs;
-- +goose StatementEnd
