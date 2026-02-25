-- +goose Up

CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    full_name TEXT,
    given_name TEXT,
    family_name TEXT,
    is_barteamer BOOLEAN DEFAULT FALSE,
    is_admin BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    spending_limit_disabled BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE categories (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE items (
    id BIGSERIAL PRIMARY KEY,
    category_id BIGINT NOT NULL REFERENCES categories(id),
    name TEXT NOT NULL,
    price_barteamer BIGINT NOT NULL,
    price_helfer BIGINT NOT NULL,
    sort_order INTEGER DEFAULT 0,
    deleted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE transactions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    amount BIGINT NOT NULL,
    item_title TEXT,
    unit_price BIGINT,
    quantity INTEGER,
    description TEXT,
    type TEXT DEFAULT 'purchase',
    cancelled_at TIMESTAMPTZ,
    cancels_id BIGINT REFERENCES transactions(id),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE settings (
    id INTEGER PRIMARY KEY DEFAULT 1,
    warning_limit BIGINT DEFAULT 0,
    hard_spending_limit BIGINT DEFAULT 0,
    hard_limit_enabled BOOLEAN DEFAULT FALSE,
    custom_tx_min BIGINT DEFAULT 0,
    custom_tx_max BIGINT DEFAULT 0,
    max_item_quantity INTEGER DEFAULT 10,
    cancellation_minutes INTEGER DEFAULT 30,
    pagination_size INTEGER DEFAULT 20,
    smtp_host TEXT DEFAULT '',
    smtp_port INTEGER DEFAULT 587,
    smtp_user TEXT DEFAULT '',
    smtp_password TEXT DEFAULT '',
    smtp_from TEXT DEFAULT '',
    email_template TEXT DEFAULT '',
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT settings_singleton CHECK (id = 1)
);

INSERT INTO settings (id) VALUES (1);

CREATE INDEX idx_transactions_user_id ON transactions(user_id);
CREATE INDEX idx_transactions_created_at ON transactions(created_at);
CREATE INDEX idx_transactions_type ON transactions(type);
CREATE INDEX idx_items_category_id ON items(category_id);
CREATE INDEX idx_items_deleted_at ON items(deleted_at) WHERE deleted_at IS NULL;

-- +goose Down

DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS items;
DROP TABLE IF EXISTS categories;
DROP TABLE IF EXISTS settings;
DROP TABLE IF EXISTS users;
