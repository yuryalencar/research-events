-- +goose Up
-- +goose StatementBegin
ALTER TABLE deadlines ADD COLUMN time VARCHAR(5);
ALTER TABLE deadlines ADD COLUMN timezone VARCHAR(50);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE deadlines DROP COLUMN time;
ALTER TABLE deadlines DROP COLUMN timezone;
-- +goose StatementEnd
