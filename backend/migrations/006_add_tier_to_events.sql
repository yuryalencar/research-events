-- +goose Up
-- +goose StatementBegin

-- Tier is the CORE-style conference ranking (A*, A, B, C), or "unranked" when
-- unknown. NOT NULL with a default so existing rows and rows inserted without
-- specifying tier are always "unranked" — never NULL.
ALTER TABLE events ADD COLUMN tier VARCHAR(20) NOT NULL DEFAULT 'unranked';

ALTER TABLE events ADD CONSTRAINT events_tier_check CHECK (tier IN ('A*', 'A', 'B', 'C', 'unranked'));

-- Supports the ?tier= filter on GET /api/v1/events.
CREATE INDEX events_tier_idx ON events (tier);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE events DROP COLUMN tier;
-- +goose StatementEnd
