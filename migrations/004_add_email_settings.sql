-- +goose Up
ALTER TABLE settings ADD COLUMN smtp_from_name TEXT NOT NULL DEFAULT '';
ALTER TABLE settings ADD COLUMN email_subject TEXT NOT NULL DEFAULT 'Kontostand-Erinnerung';

-- +goose Down
ALTER TABLE settings DROP COLUMN IF EXISTS email_subject;
ALTER TABLE settings DROP COLUMN IF EXISTS smtp_from_name;
