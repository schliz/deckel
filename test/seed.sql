-- Seed data for e2e tests.
-- Executed by Makefile after the app has applied Goose migrations.

-- Update the settings singleton with test defaults
UPDATE settings SET
    warning_limit       = -1000,
    hard_spending_limit = 2000,
    hard_limit_enabled  = true,
    custom_tx_min       = -500,
    custom_tx_max       = 500,
    max_item_quantity   = 10,
    cancellation_minutes = 30,
    pagination_size     = 20
WHERE id = 1;

-- Sample menu: two categories with items
INSERT INTO categories (id, name, sort_order) VALUES
    (1, 'Getränke',  1),
    (2, 'Snacks',    2)
ON CONFLICT (id) DO NOTHING;

INSERT INTO items (id, category_id, name, price_barteamer, price_helfer, sort_order) VALUES
    (1, 1, 'Bier',      150, 200, 1),
    (2, 1, 'Mate',      150, 200, 2),
    (3, 2, 'Chips',     100, 150, 1)
ON CONFLICT (id) DO NOTHING;

-- Reset sequences so future inserts don't collide
SELECT setval('categories_id_seq', (SELECT COALESCE(MAX(id), 0) FROM categories));
SELECT setval('items_id_seq',      (SELECT COALESCE(MAX(id), 0) FROM items));
