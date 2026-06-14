-- +goose Up
-- +goose StatementBegin
ALTER TABLE audit_logs ADD COLUMN reason TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE audit_logs DROP COLUMN reason;
-- +goose StatementEnd
