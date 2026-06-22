-- +goose Up
-- +goose StatementBegin

-- Seed default admin and moderator accounts used as fixtures for manual
-- and automated testing (login, JWT, admin review flows, etc.).
-- Passwords are bcrypt-hashed at cost 12, per specs/backend/auth-login.yaml.
--
-- Admin    -> yuryalencar19@gmail.com / ReSearchv1.0@dmin
-- Moderator -> moderator@example.com  / Moderator123!
--
-- ON CONFLICT DO NOTHING keeps this migration idempotent and safe to apply
-- to any environment without overwriting an existing account with the same email.
INSERT INTO users (name, email, password_hash, role, created_at, updated_at)
VALUES
    ('Yury Lima', 'yuryalencar19@gmail.com', '$2a$12$.chn8CW74eRM/uE9.S1I9eYZR.jt0CdPLuAPclVKBHZDOVHqQ44pW', 'admin', NOW(), NOW()),
    ('Default Moderator', 'moderator@example.com', '$2a$12$DoAkpeLQwOrDtejOHxZJhu0dTbjGQVd6VETmIbzzG/TpcVBT.D0Ba', 'moderator', NOW(), NOW())
ON CONFLICT (email) DO NOTHING;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM users WHERE email IN ('yuryalencar19@gmail.com', 'moderator@example.com');
-- +goose StatementEnd
