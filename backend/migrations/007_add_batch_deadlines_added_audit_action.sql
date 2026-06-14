-- +goose Up
-- +goose StatementBegin
ALTER TABLE audit_logs DROP CONSTRAINT audit_logs_action_check;
ALTER TABLE audit_logs ADD CONSTRAINT audit_logs_action_check CHECK (action IN (
    'created', 'updated', 'approved', 'rejected',
    'deadline_added', 'deadline_superseded', 'unlocked',
    'batch_deadlines_added'
));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE audit_logs DROP CONSTRAINT audit_logs_action_check;
ALTER TABLE audit_logs ADD CONSTRAINT audit_logs_action_check CHECK (action IN (
    'created', 'updated', 'approved', 'rejected',
    'deadline_added', 'deadline_superseded', 'unlocked'
));
-- +goose StatementEnd
