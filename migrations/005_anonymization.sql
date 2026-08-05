-- +goose Up
ALTER TABLE transactions ALTER COLUMN user_id DROP NOT NULL;
ALTER TABLE transactions ADD COLUMN anonymized_at TIMESTAMPTZ;
ALTER TABLE transactions ADD CONSTRAINT transactions_user_or_anonymized
    CHECK (user_id IS NOT NULL OR anonymized_at IS NOT NULL);
ALTER TABLE users ADD COLUMN balance_carryover BIGINT NOT NULL DEFAULT 0;
ALTER TABLE settings ADD COLUMN anonymize_after_months INTEGER NOT NULL DEFAULT 36;
CREATE TABLE anonymization_runs (
    id BIGSERIAL PRIMARY KEY,
    executed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    cutoff_date TIMESTAMPTZ NOT NULL,
    transaction_count INTEGER NOT NULL,
    user_count INTEGER NOT NULL,
    executed_by TEXT NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS anonymization_runs;
ALTER TABLE settings DROP COLUMN IF EXISTS anonymize_after_months;
ALTER TABLE users DROP COLUMN IF EXISTS balance_carryover;
ALTER TABLE transactions DROP CONSTRAINT IF EXISTS transactions_user_or_anonymized;
ALTER TABLE transactions DROP COLUMN IF EXISTS anonymized_at;
ALTER TABLE transactions ALTER COLUMN user_id SET NOT NULL;
