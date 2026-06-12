-- +goose Up
-- +goose StatementBegin
CREATE TABLE events (
    id                 BIGSERIAL    PRIMARY KEY,
    name               VARCHAR(255) NOT NULL,
    slug               VARCHAR(255) NOT NULL,
    country            VARCHAR(255) NOT NULL,
    city               VARCHAR(255) NOT NULL,
    latitude           DOUBLE PRECISION NOT NULL,
    longitude          DOUBLE PRECISION NOT NULL,
    start_date         TIMESTAMPTZ  NOT NULL,
    end_date           TIMESTAMPTZ  NOT NULL,
    website_url        VARCHAR(2048) NOT NULL,

    -- domain is an intentionally open string enum (no CHECK constraint) so new
    -- domains can be added without a migration; validated in the service layer.
    domain             VARCHAR(100) NOT NULL,

    status             VARCHAR(50)  NOT NULL DEFAULT 'pending',

    -- Indexed for the default "current year" filter on the globe homepage.
    year               INTEGER      NOT NULL,

    created_by_id      BIGINT       NOT NULL REFERENCES users(id),
    last_updated_by_id BIGINT       NOT NULL REFERENCES users(id),

    -- GORM-compatible timestamps (soft delete via deleted_at).
    created_at         TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at         TIMESTAMPTZ,

    CONSTRAINT events_status_check CHECK (status IN ('pending', 'approved', 'rejected'))
);

-- Slug must be unique among pending/approved events only — a rejected event's
-- slug can be reused by a new submission, so the partial index excludes
-- status='rejected' (and soft-deleted rows) from the uniqueness check.
CREATE UNIQUE INDEX events_slug_idx ON events (slug)
    WHERE deleted_at IS NULL AND status IN ('pending', 'approved');

-- Bounding-box queries for the globe viewport filter.
CREATE INDEX events_lat_lng_idx ON events (latitude, longitude);

-- Year filter (default view = current year) and domain filter.
CREATE INDEX events_year_idx ON events (year);
CREATE INDEX events_domain_idx ON events (domain);
CREATE INDEX events_status_idx ON events (status);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS events;
-- +goose StatementEnd
