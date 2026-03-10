-- +goose Up
ALTER TABLE transactions ADD COLUMN created_by_user_id BIGINT REFERENCES users(id);
CREATE INDEX idx_transactions_created_by ON transactions(created_by_user_id) WHERE created_by_user_id IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_transactions_created_by;
ALTER TABLE transactions DROP COLUMN IF EXISTS created_by_user_id;
