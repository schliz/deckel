-- Seed data for e2e tests.
-- Executed by Makefile after the app has applied Goose migrations.

-- Update the settings singleton with test defaults
UPDATE settings SET
    warning_limit         = -1000,
    hard_spending_limit   = 2000,
    hard_limit_enabled    = true,
    custom_tx_min         = -500,
    custom_tx_max         = 500,
    max_item_quantity     = 10,
    cancellation_minutes  = 30,
    pagination_size       = 20,
    anonymize_after_months = 36
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

-- Anonymization + user-deletion test fixtures (see docs/anonymisierung.md).
--
-- User 901 "Old Anon User": active, with four 37-month-old transactions
-- (older than the 36-month anonymize_after_months cutoff set above) plus
-- one fresh transaction that must stay untouched by a run.
-- User 902 "Delete Me": deactivated, no transactions -> deletable.
-- User 903 "Delete Blocked": deactivated, one fresh transaction -> deletion
-- must be blocked until that transaction is anonymized.
INSERT INTO users (id, email, full_name, given_name, family_name, is_active) VALUES
    (901, 'anon-old@test.example',    'Old Anon User',  'Old',    'Anon User', false),
    (902, 'delete-me@test.example',   'Delete Me',      'Delete', 'Me',        false),
    (903, 'delete-block@test.example','Delete Blocked', 'Delete', 'Blocked',   false)
ON CONFLICT (id) DO NOTHING;

-- Old transactions for user 901 (37 months old): deposit, a purchase, a
-- cancelled purchase and its cancellation counter-transaction. Both halves
-- of the cancellation pair are equally old, so the pair is anonymized
-- together (see cancels_id pairing rule).
INSERT INTO transactions
    (id, user_id, amount, item_title, unit_price, quantity, description, type, cancelled_at, cancels_id, created_by_user_id, created_at) VALUES
    (901, 901,  1000, NULL,   NULL, NULL, 'Einzahlung',   'deposit',     NULL,                          NULL, 901, NOW() - INTERVAL '37 months'),
    (902, 901,  -300, 'Bier', 150,  2,    NULL,           'purchase',    NULL,                          NULL, 901, NOW() - INTERVAL '37 months'),
    (903, 901,  -150, 'Mate', 150,  1,    NULL,           'purchase',    NOW() - INTERVAL '37 months',  NULL, 901, NOW() - INTERVAL '37 months'),
    (904, 901,   150, NULL,   NULL, NULL, 'Storno: Mate', 'cancellation', NULL,                         903,  901, NOW() - INTERVAL '37 months')
ON CONFLICT (id) DO NOTHING;

-- Fresh purchase for user 901: must remain untouched (not older than the
-- cutoff), keeping the balance and anonymization preview predictable.
-- Balance user 901: +1000 - 300 - 150 + 150 - 200 = +500 cents (5,00 EUR).
INSERT INTO transactions
    (id, user_id, amount, item_title, unit_price, quantity, description, type, created_by_user_id, created_at) VALUES
    (905, 901, -200, 'Bier', 200, 1, NULL, 'purchase', 901, NOW())
ON CONFLICT (id) DO NOTHING;

-- Fresh custom transaction for user 903: blocks deletion until anonymized.
INSERT INTO transactions
    (id, user_id, amount, description, type, created_by_user_id, created_at) VALUES
    (906, 903, -100, 'Blocker', 'custom', 903, NOW())
ON CONFLICT (id) DO NOTHING;

SELECT setval('users_id_seq',        (SELECT COALESCE(MAX(id), 0) FROM users));
SELECT setval('transactions_id_seq', (SELECT COALESCE(MAX(id), 0) FROM transactions));
