-- Demo member data export (users + transactions), generated 2026-08-05.
-- Self-contained, re-runnable SQL transaction.
--
-- Transaction timestamps are stored below as offsets from a fixed anchor
-- date (the export date, 2026-08-05) and are re-anchored to CURRENT_DATE at
-- import time (see `CURRENT_DATE::timestamptz + INTERVAL '...'` below), so
-- re-importing this file at any later date reproduces the same relative
-- spread of dates (e.g. "-1952 days" before today) rather than the
-- original absolute dates.
--
-- Idempotent: each user's transaction batch is skipped if that user
-- already has any transactions, so re-running this file is safe.

BEGIN;

-- Upsert demo users
INSERT INTO users (email, full_name, given_name, family_name, is_active) VALUES
    ('anna.bauer@example.com', 'Anna Bauer', 'Anna', 'Bauer', true),
    ('jonas.weber@example.com', 'Jonas Weber', 'Jonas', 'Weber', true),
    ('lea.schmidt@example.com', 'Lea Schmidt', 'Lea', 'Schmidt', true),
    ('tim.hoffmann@example.com', 'Tim Hoffmann', 'Tim', 'Hoffmann', true),
    ('nina.krueger@example.com', 'Nina Krueger', 'Nina', 'Krueger', false),
    ('paul.richter@example.com', 'Paul Richter', 'Paul', 'Richter', false)
ON CONFLICT (email) DO UPDATE SET email = EXCLUDED.email;

-- Transactions for Anna Bauer (anna.bauer@example.com) -- 60 rows
DO $$
DECLARE
    uid BIGINT;
    already_seeded BOOLEAN;
BEGIN
    SELECT id INTO uid FROM users WHERE email = 'anna.bauer@example.com';
    SELECT EXISTS(SELECT 1 FROM transactions WHERE user_id = uid) INTO already_seeded;
    IF NOT already_seeded THEN
        INSERT INTO transactions
            (user_id, amount, item_title, unit_price, quantity, description, type, created_by_user_id, created_at)
        VALUES
            (( SELECT id FROM users WHERE email='anna.bauer@example.com' ), '9000', NULL, NULL, NULL, 'Einzahlung', 'deposit', ( SELECT id FROM users WHERE email='bark4-kasse@hadiko.de' ), CURRENT_DATE::timestamptz + INTERVAL '-1952 days -12:05:10.676433'),
            (( SELECT id FROM users WHERE email='anna.bauer@example.com' ), '9500', NULL, NULL, NULL, 'Einzahlung', 'deposit', ( SELECT id FROM users WHERE email='bark4-kasse@hadiko.de' ), CURRENT_DATE::timestamptz + INTERVAL '-696 days -21:57:17.69127'),
            (( SELECT id FROM users WHERE email='anna.bauer@example.com' ), '4500', NULL, NULL, NULL, 'Einzahlung', 'deposit', ( SELECT id FROM users WHERE email='bark4-kasse@hadiko.de' ), CURRENT_DATE::timestamptz + INTERVAL '-1327 days -19:57:14.133186'),
            (( SELECT id FROM users WHERE email='anna.bauer@example.com' ), '-717', NULL, NULL, NULL, 'Sonderbuchung', 'custom', ( SELECT id FROM users WHERE email='bark4-kasse@hadiko.de' ), CURRENT_DATE::timestamptz + INTERVAL '-314 days -14:04:21.299479'),
            (( SELECT id FROM users WHERE email='anna.bauer@example.com' ), '-150', 'Flaschenbier 0,5 L', '150', '1', NULL, 'purchase', ( SELECT id FROM users WHERE email='anna.bauer@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-1884 days -04:16:10.342505'),
            (( SELECT id FROM users WHERE email='anna.bauer@example.com' ), '9000', NULL, NULL, NULL, 'Einzahlung', 'deposit', ( SELECT id FROM users WHERE email='bark4-kasse@hadiko.de' ), CURRENT_DATE::timestamptz + INTERVAL '-2237 days -01:42:47.671936'),
            (( SELECT id FROM users WHERE email='anna.bauer@example.com' ), '-450', 'Flaschenbier 0,5 L', '150', '3', NULL, 'purchase', ( SELECT id FROM users WHERE email='anna.bauer@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-540 days -19:13:36.970177'),
            (( SELECT id FROM users WHERE email='anna.bauer@example.com' ), '-150', 'Flaschenbier 0,5 L', '150', '1', NULL, 'purchase', ( SELECT id FROM users WHERE email='anna.bauer@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-1071 days -22:10:25.195155'),
            (( SELECT id FROM users WHERE email='anna.bauer@example.com' ), '4500', NULL, NULL, NULL, 'Einzahlung', 'deposit', ( SELECT id FROM users WHERE email='bark4-kasse@hadiko.de' ), CURRENT_DATE::timestamptz + INTERVAL '-922 days -23:23:32.999226'),
            (( SELECT id FROM users WHERE email='anna.bauer@example.com' ), '-330', 'Flaschenbier 0,33 L', '110', '3', NULL, 'purchase', ( SELECT id FROM users WHERE email='anna.bauer@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-09:29:46.078989'),
            (( SELECT id FROM users WHERE email='anna.bauer@example.com' ), '-360', 'Fassbier Helles', '180', '2', NULL, 'purchase', ( SELECT id FROM users WHERE email='anna.bauer@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-2188 days -15:16:52.739022'),
            (( SELECT id FROM users WHERE email='anna.bauer@example.com' ), '-180', 'Fassbier Helles', '180', '1', NULL, 'purchase', ( SELECT id FROM users WHERE email='anna.bauer@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-1162 days -02:05:10.874061'),
            (( SELECT id FROM users WHERE email='anna.bauer@example.com' ), '-110', 'Flaschenbier 0,33 L', '110', '1', NULL, 'purchase', ( SELECT id FROM users WHERE email='anna.bauer@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-1260 days -14:11:52.92807'),
            (( SELECT id FROM users WHERE email='anna.bauer@example.com' ), '6000', NULL, NULL, NULL, 'Einzahlung', 'deposit', ( SELECT id FROM users WHERE email='bark4-kasse@hadiko.de' ), CURRENT_DATE::timestamptz + INTERVAL '-318 days -15:32:36.138465'),
            (( SELECT id FROM users WHERE email='anna.bauer@example.com' ), '-540', 'Fassbier Helles', '180', '3', NULL, 'purchase', ( SELECT id FROM users WHERE email='anna.bauer@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-198 days -04:42:01.940511'),
            (( SELECT id FROM users WHERE email='anna.bauer@example.com' ), '-360', 'Fassbier Helles', '180', '2', NULL, 'purchase', ( SELECT id FROM users WHERE email='anna.bauer@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-1859 days -14:10:01.111002'),
            (( SELECT id FROM users WHERE email='anna.bauer@example.com' ), '11000', NULL, NULL, NULL, 'Einzahlung', 'deposit', ( SELECT id FROM users WHERE email='bark4-kasse@hadiko.de' ), CURRENT_DATE::timestamptz + INTERVAL '-49 days -20:38:40.652051'),
            (( SELECT id FROM users WHERE email='anna.bauer@example.com' ), '8500', NULL, NULL, NULL, 'Einzahlung', 'deposit', ( SELECT id FROM users WHERE email='bark4-kasse@hadiko.de' ), CURRENT_DATE::timestamptz + INTERVAL '-1554 days -08:51:29.018504'),
            (( SELECT id FROM users WHERE email='anna.bauer@example.com' ), '-110', 'Flaschenbier 0,33 L', '110', '1', NULL, 'purchase', ( SELECT id FROM users WHERE email='anna.bauer@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-2008 days -07:21:50.522818'),
            (( SELECT id FROM users WHERE email='anna.bauer@example.com' ), '8500', NULL, NULL, NULL, 'Einzahlung', 'deposit', ( SELECT id FROM users WHERE email='bark4-kasse@hadiko.de' ), CURRENT_DATE::timestamptz + INTERVAL '-359 days -12:58:24.762086'),
            (( SELECT id FROM users WHERE email='anna.bauer@example.com' ), '8500', NULL, NULL, NULL, 'Einzahlung', 'deposit', ( SELECT id FROM users WHERE email='bark4-kasse@hadiko.de' ), CURRENT_DATE::timestamptz + INTERVAL '-1435 days -15:45:49.276828'),
            (( SELECT id FROM users WHERE email='anna.bauer@example.com' ), '-161', NULL, NULL, NULL, 'Sonderbuchung', 'custom', ( SELECT id FROM users WHERE email='bark4-kasse@hadiko.de' ), CURRENT_DATE::timestamptz + INTERVAL '-2344 days -04:29:57.834688'),
            (( SELECT id FROM users WHERE email='anna.bauer@example.com' ), '-150', 'Flaschenbier 0,5 L', '150', '1', NULL, 'purchase', ( SELECT id FROM users WHERE email='anna.bauer@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-775 days -02:10:27.675977'),
            (( SELECT id FROM users WHERE email='anna.bauer@example.com' ), '6000', NULL, NULL, NULL, 'Einzahlung', 'deposit', ( SELECT id FROM users WHERE email='bark4-kasse@hadiko.de' ), CURRENT_DATE::timestamptz + INTERVAL '-2152 days -19:09:15.870675'),
            (( SELECT id FROM users WHERE email='anna.bauer@example.com' ), '-540', 'Fassbier Helles', '180', '3', NULL, 'purchase', ( SELECT id FROM users WHERE email='anna.bauer@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-2274 days -04:27:26.896371'),
            (( SELECT id FROM users WHERE email='anna.bauer@example.com' ), '-330', 'Flaschenbier 0,33 L', '110', '3', NULL, 'purchase', ( SELECT id FROM users WHERE email='anna.bauer@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-2128 days -18:56:21.893233'),
            (( SELECT id FROM users WHERE email='anna.bauer@example.com' ), '-450', 'Flaschenbier 0,5 L', '150', '3', NULL, 'purchase', ( SELECT id FROM users WHERE email='anna.bauer@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-1371 days -14:57:32.797909'),
            (( SELECT id FROM users WHERE email='anna.bauer@example.com' ), '-360', 'Fassbier Helles', '180', '2', NULL, 'purchase', ( SELECT id FROM users WHERE email='anna.bauer@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-2358 days -16:23:21.97057'),
            (( SELECT id FROM users WHERE email='anna.bauer@example.com' ), '-220', 'Flaschenbier 0,33 L', '110', '2', NULL, 'purchase', ( SELECT id FROM users WHERE email='anna.bauer@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-746 days -18:52:37.026031'),
            (( SELECT id FROM users WHERE email='anna.bauer@example.com' ), '-150', 'Flaschenbier 0,5 L', '150', '1', NULL, 'purchase', ( SELECT id FROM users WHERE email='anna.bauer@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-1940 days -13:37:45.528057'),
            (( SELECT id FROM users WHERE email='anna.bauer@example.com' ), '-330', 'Flaschenbier 0,33 L', '110', '3', NULL, 'purchase', ( SELECT id FROM users WHERE email='anna.bauer@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-491 days -14:45:53.951988'),
            (( SELECT id FROM users WHERE email='anna.bauer@example.com' ), '-220', 'Flaschenbier 0,33 L', '110', '2', NULL, 'purchase', ( SELECT id FROM users WHERE email='anna.bauer@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-2337 days -00:13:56.222058'),
            (( SELECT id FROM users WHERE email='anna.bauer@example.com' ), '448', NULL, NULL, NULL, 'Sonderbuchung', 'custom', ( SELECT id FROM users WHERE email='bark4-kasse@hadiko.de' ), CURRENT_DATE::timestamptz + INTERVAL '-460 days -16:50:38.969211'),
            (( SELECT id FROM users WHERE email='anna.bauer@example.com' ), '6000', NULL, NULL, NULL, 'Einzahlung', 'deposit', ( SELECT id FROM users WHERE email='bark4-kasse@hadiko.de' ), CURRENT_DATE::timestamptz + INTERVAL '-1441 days -03:39:54.797828'),
            (( SELECT id FROM users WHERE email='anna.bauer@example.com' ), '-110', 'Flaschenbier 0,33 L', '110', '1', NULL, 'purchase', ( SELECT id FROM users WHERE email='anna.bauer@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-192 days -01:01:21.959534'),
            (( SELECT id FROM users WHERE email='anna.bauer@example.com' ), '-300', 'Flaschenbier 0,5 L', '150', '2', NULL, 'purchase', ( SELECT id FROM users WHERE email='anna.bauer@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-1680 days -17:02:12.552435'),
            (( SELECT id FROM users WHERE email='anna.bauer@example.com' ), '-180', 'Fassbier Helles', '180', '1', NULL, 'purchase', ( SELECT id FROM users WHERE email='anna.bauer@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-1253 days -15:43:00.014211'),
            (( SELECT id FROM users WHERE email='anna.bauer@example.com' ), '-330', 'Flaschenbier 0,33 L', '110', '3', NULL, 'purchase', ( SELECT id FROM users WHERE email='anna.bauer@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-175 days -07:32:50.262521'),
            (( SELECT id FROM users WHERE email='anna.bauer@example.com' ), '8500', NULL, NULL, NULL, 'Einzahlung', 'deposit', ( SELECT id FROM users WHERE email='bark4-kasse@hadiko.de' ), CURRENT_DATE::timestamptz + INTERVAL '-291 days -22:22:31.736022'),
            (( SELECT id FROM users WHERE email='anna.bauer@example.com' ), '-450', 'Flaschenbier 0,5 L', '150', '3', NULL, 'purchase', ( SELECT id FROM users WHERE email='anna.bauer@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-2246 days -17:16:21.152751'),
            (( SELECT id FROM users WHERE email='anna.bauer@example.com' ), '-150', 'Flaschenbier 0,5 L', '150', '1', NULL, 'purchase', ( SELECT id FROM users WHERE email='anna.bauer@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-1662 days -20:02:31.695278'),
            (( SELECT id FROM users WHERE email='anna.bauer@example.com' ), '-330', 'Flaschenbier 0,33 L', '110', '3', NULL, 'purchase', ( SELECT id FROM users WHERE email='anna.bauer@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-58 days -15:43:38.229642'),
            (( SELECT id FROM users WHERE email='anna.bauer@example.com' ), '-110', 'Flaschenbier 0,33 L', '110', '1', NULL, 'purchase', ( SELECT id FROM users WHERE email='anna.bauer@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-183 days -18:58:49.687247'),
            (( SELECT id FROM users WHERE email='anna.bauer@example.com' ), '-180', 'Fassbier Helles', '180', '1', NULL, 'purchase', ( SELECT id FROM users WHERE email='anna.bauer@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-1909 days -04:16:02.388266'),
            (( SELECT id FROM users WHERE email='anna.bauer@example.com' ), '-300', 'Flaschenbier 0,5 L', '150', '2', NULL, 'purchase', ( SELECT id FROM users WHERE email='anna.bauer@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-491 days -09:28:35.380401'),
            (( SELECT id FROM users WHERE email='anna.bauer@example.com' ), '-450', 'Flaschenbier 0,5 L', '150', '3', NULL, 'purchase', ( SELECT id FROM users WHERE email='anna.bauer@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-1328 days -15:13:50.598482'),
            (( SELECT id FROM users WHERE email='anna.bauer@example.com' ), '6500', NULL, NULL, NULL, 'Einzahlung', 'deposit', ( SELECT id FROM users WHERE email='bark4-kasse@hadiko.de' ), CURRENT_DATE::timestamptz + INTERVAL '-199 days -22:11:30.333676'),
            (( SELECT id FROM users WHERE email='anna.bauer@example.com' ), '6500', NULL, NULL, NULL, 'Einzahlung', 'deposit', ( SELECT id FROM users WHERE email='bark4-kasse@hadiko.de' ), CURRENT_DATE::timestamptz + INTERVAL '-1791 days -07:06:20.400354'),
            (( SELECT id FROM users WHERE email='anna.bauer@example.com' ), '4000', NULL, NULL, NULL, 'Einzahlung', 'deposit', ( SELECT id FROM users WHERE email='bark4-kasse@hadiko.de' ), CURRENT_DATE::timestamptz + INTERVAL '-1265 days -09:19:30.793798'),
            (( SELECT id FROM users WHERE email='anna.bauer@example.com' ), '-220', 'Flaschenbier 0,33 L', '110', '2', NULL, 'purchase', ( SELECT id FROM users WHERE email='anna.bauer@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-1561 days -11:00:07.059313'),
            (( SELECT id FROM users WHERE email='anna.bauer@example.com' ), '10500', NULL, NULL, NULL, 'Einzahlung', 'deposit', ( SELECT id FROM users WHERE email='bark4-kasse@hadiko.de' ), CURRENT_DATE::timestamptz + INTERVAL '-1699 days -14:58:32.5884'),
            (( SELECT id FROM users WHERE email='anna.bauer@example.com' ), '-180', 'Fassbier Helles', '180', '1', NULL, 'purchase', ( SELECT id FROM users WHERE email='anna.bauer@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-617 days -10:19:51.640006'),
            (( SELECT id FROM users WHERE email='anna.bauer@example.com' ), '-220', 'Flaschenbier 0,33 L', '110', '2', NULL, 'purchase', ( SELECT id FROM users WHERE email='anna.bauer@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-2330 days -01:23:33.976083'),
            (( SELECT id FROM users WHERE email='anna.bauer@example.com' ), '-450', 'Flaschenbier 0,5 L', '150', '3', NULL, 'purchase', ( SELECT id FROM users WHERE email='anna.bauer@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-91 days -07:56:54.734729'),
            (( SELECT id FROM users WHERE email='anna.bauer@example.com' ), '-300', 'Flaschenbier 0,5 L', '150', '2', NULL, 'purchase', ( SELECT id FROM users WHERE email='anna.bauer@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-1119 days -17:34:50.095827'),
            (( SELECT id FROM users WHERE email='anna.bauer@example.com' ), '-300', 'Flaschenbier 0,5 L', '150', '2', NULL, 'purchase', ( SELECT id FROM users WHERE email='anna.bauer@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-851 days -22:21:18.046064'),
            (( SELECT id FROM users WHERE email='anna.bauer@example.com' ), '-180', 'Fassbier Helles', '180', '1', NULL, 'purchase', ( SELECT id FROM users WHERE email='anna.bauer@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-1946 days -04:35:24.687335'),
            (( SELECT id FROM users WHERE email='anna.bauer@example.com' ), '10500', NULL, NULL, NULL, 'Einzahlung', 'deposit', ( SELECT id FROM users WHERE email='bark4-kasse@hadiko.de' ), CURRENT_DATE::timestamptz + INTERVAL '-216 days -17:01:03.671732'),
            (( SELECT id FROM users WHERE email='anna.bauer@example.com' ), '-360', 'Fassbier Helles', '180', '2', NULL, 'purchase', ( SELECT id FROM users WHERE email='anna.bauer@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-2249 days -09:03:43.330838'),
            (( SELECT id FROM users WHERE email='anna.bauer@example.com' ), '109', NULL, NULL, NULL, 'Sonderbuchung', 'custom', ( SELECT id FROM users WHERE email='bark4-kasse@hadiko.de' ), CURRENT_DATE::timestamptz + INTERVAL '-1950 days -14:49:54.52022');
    END IF;
END $$;

-- Transactions for Jonas Weber (jonas.weber@example.com) -- 55 rows
DO $$
DECLARE
    uid BIGINT;
    already_seeded BOOLEAN;
BEGIN
    SELECT id INTO uid FROM users WHERE email = 'jonas.weber@example.com';
    SELECT EXISTS(SELECT 1 FROM transactions WHERE user_id = uid) INTO already_seeded;
    IF NOT already_seeded THEN
        INSERT INTO transactions
            (user_id, amount, item_title, unit_price, quantity, description, type, created_by_user_id, created_at)
        VALUES
            (( SELECT id FROM users WHERE email='jonas.weber@example.com' ), '-540', 'Fassbier Helles', '180', '3', NULL, 'purchase', ( SELECT id FROM users WHERE email='jonas.weber@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-764 days -09:19:05.658197'),
            (( SELECT id FROM users WHERE email='jonas.weber@example.com' ), '-450', 'Flaschenbier 0,5 L', '150', '3', NULL, 'purchase', ( SELECT id FROM users WHERE email='jonas.weber@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-744 days -00:15:15.522002'),
            (( SELECT id FROM users WHERE email='jonas.weber@example.com' ), '-540', 'Fassbier Helles', '180', '3', NULL, 'purchase', ( SELECT id FROM users WHERE email='jonas.weber@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-419 days -09:31:03.411263'),
            (( SELECT id FROM users WHERE email='jonas.weber@example.com' ), '-150', 'Flaschenbier 0,5 L', '150', '1', NULL, 'purchase', ( SELECT id FROM users WHERE email='jonas.weber@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-1135 days -14:44:35.691497'),
            (( SELECT id FROM users WHERE email='jonas.weber@example.com' ), '-180', 'Fassbier Helles', '180', '1', NULL, 'purchase', ( SELECT id FROM users WHERE email='jonas.weber@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-2032 days -07:05:26.627285'),
            (( SELECT id FROM users WHERE email='jonas.weber@example.com' ), '-110', 'Flaschenbier 0,33 L', '110', '1', NULL, 'purchase', ( SELECT id FROM users WHERE email='jonas.weber@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-409 days -11:08:38.838109'),
            (( SELECT id FROM users WHERE email='jonas.weber@example.com' ), '-150', 'Flaschenbier 0,5 L', '150', '1', NULL, 'purchase', ( SELECT id FROM users WHERE email='jonas.weber@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-819 days -13:12:46.517127'),
            (( SELECT id FROM users WHERE email='jonas.weber@example.com' ), '-540', 'Fassbier Helles', '180', '3', NULL, 'purchase', ( SELECT id FROM users WHERE email='jonas.weber@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-869 days -16:39:15.890088'),
            (( SELECT id FROM users WHERE email='jonas.weber@example.com' ), '-540', 'Fassbier Helles', '180', '3', NULL, 'purchase', ( SELECT id FROM users WHERE email='jonas.weber@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-1287 days -10:37:03.665138'),
            (( SELECT id FROM users WHERE email='jonas.weber@example.com' ), '-150', 'Flaschenbier 0,5 L', '150', '1', NULL, 'purchase', ( SELECT id FROM users WHERE email='jonas.weber@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-735 days -10:22:03.622922'),
            (( SELECT id FROM users WHERE email='jonas.weber@example.com' ), '-110', 'Flaschenbier 0,33 L', '110', '1', NULL, 'purchase', ( SELECT id FROM users WHERE email='jonas.weber@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-1195 days -09:19:31.970972'),
            (( SELECT id FROM users WHERE email='jonas.weber@example.com' ), '-150', 'Flaschenbier 0,5 L', '150', '1', NULL, 'purchase', ( SELECT id FROM users WHERE email='jonas.weber@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-769 days -15:11:40.238904'),
            (( SELECT id FROM users WHERE email='jonas.weber@example.com' ), '-330', 'Flaschenbier 0,33 L', '110', '3', NULL, 'purchase', ( SELECT id FROM users WHERE email='jonas.weber@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-873 days -03:15:40.468965'),
            (( SELECT id FROM users WHERE email='jonas.weber@example.com' ), '-220', 'Flaschenbier 0,33 L', '110', '2', NULL, 'purchase', ( SELECT id FROM users WHERE email='jonas.weber@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-1568 days -19:12:18.241916'),
            (( SELECT id FROM users WHERE email='jonas.weber@example.com' ), '-220', 'Flaschenbier 0,33 L', '110', '2', NULL, 'purchase', ( SELECT id FROM users WHERE email='jonas.weber@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-2277 days -04:40:54.585834'),
            (( SELECT id FROM users WHERE email='jonas.weber@example.com' ), '-220', 'Flaschenbier 0,33 L', '110', '2', NULL, 'purchase', ( SELECT id FROM users WHERE email='jonas.weber@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-2085 days -14:44:04.457271'),
            (( SELECT id FROM users WHERE email='jonas.weber@example.com' ), '7000', NULL, NULL, NULL, 'Einzahlung', 'deposit', ( SELECT id FROM users WHERE email='bark4-kasse@hadiko.de' ), CURRENT_DATE::timestamptz + INTERVAL '-1076 days -10:26:33.460477'),
            (( SELECT id FROM users WHERE email='jonas.weber@example.com' ), '-330', 'Flaschenbier 0,33 L', '110', '3', NULL, 'purchase', ( SELECT id FROM users WHERE email='jonas.weber@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-1948 days -22:56:00.344808'),
            (( SELECT id FROM users WHERE email='jonas.weber@example.com' ), '-450', 'Flaschenbier 0,5 L', '150', '3', NULL, 'purchase', ( SELECT id FROM users WHERE email='jonas.weber@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-1060 days -00:10:31.125467'),
            (( SELECT id FROM users WHERE email='jonas.weber@example.com' ), '-540', 'Fassbier Helles', '180', '3', NULL, 'purchase', ( SELECT id FROM users WHERE email='jonas.weber@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-1419 days -04:38:53.585479'),
            (( SELECT id FROM users WHERE email='jonas.weber@example.com' ), '-110', 'Flaschenbier 0,33 L', '110', '1', NULL, 'purchase', ( SELECT id FROM users WHERE email='jonas.weber@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-2176 days -20:57:46.584025'),
            (( SELECT id FROM users WHERE email='jonas.weber@example.com' ), '4000', NULL, NULL, NULL, 'Einzahlung', 'deposit', ( SELECT id FROM users WHERE email='bark4-kasse@hadiko.de' ), CURRENT_DATE::timestamptz + INTERVAL '-902 days -19:00:44.936111'),
            (( SELECT id FROM users WHERE email='jonas.weber@example.com' ), '-450', 'Flaschenbier 0,5 L', '150', '3', NULL, 'purchase', ( SELECT id FROM users WHERE email='jonas.weber@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-552 days -21:19:57.895837'),
            (( SELECT id FROM users WHERE email='jonas.weber@example.com' ), '-220', 'Flaschenbier 0,33 L', '110', '2', NULL, 'purchase', ( SELECT id FROM users WHERE email='jonas.weber@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-757 days -09:46:35.814424'),
            (( SELECT id FROM users WHERE email='jonas.weber@example.com' ), '-330', 'Flaschenbier 0,33 L', '110', '3', NULL, 'purchase', ( SELECT id FROM users WHERE email='jonas.weber@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-1188 days -16:24:09.76895'),
            (( SELECT id FROM users WHERE email='jonas.weber@example.com' ), '-360', 'Fassbier Helles', '180', '2', NULL, 'purchase', ( SELECT id FROM users WHERE email='jonas.weber@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-1403 days -15:07:39.310899'),
            (( SELECT id FROM users WHERE email='jonas.weber@example.com' ), '-150', 'Flaschenbier 0,5 L', '150', '1', NULL, 'purchase', ( SELECT id FROM users WHERE email='jonas.weber@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-2037 days -21:33:03.022784'),
            (( SELECT id FROM users WHERE email='jonas.weber@example.com' ), '-300', 'Flaschenbier 0,5 L', '150', '2', NULL, 'purchase', ( SELECT id FROM users WHERE email='jonas.weber@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-1327 days -22:33:27.051893'),
            (( SELECT id FROM users WHERE email='jonas.weber@example.com' ), '-150', 'Flaschenbier 0,5 L', '150', '1', NULL, 'purchase', ( SELECT id FROM users WHERE email='jonas.weber@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-230 days -04:22:32.763151'),
            (( SELECT id FROM users WHERE email='jonas.weber@example.com' ), '8500', NULL, NULL, NULL, 'Einzahlung', 'deposit', ( SELECT id FROM users WHERE email='bark4-kasse@hadiko.de' ), CURRENT_DATE::timestamptz + INTERVAL '-265 days -06:03:48.512697'),
            (( SELECT id FROM users WHERE email='jonas.weber@example.com' ), '-540', 'Fassbier Helles', '180', '3', NULL, 'purchase', ( SELECT id FROM users WHERE email='jonas.weber@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-480 days -07:32:10.935425'),
            (( SELECT id FROM users WHERE email='jonas.weber@example.com' ), '-110', 'Flaschenbier 0,33 L', '110', '1', NULL, 'purchase', ( SELECT id FROM users WHERE email='jonas.weber@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-2006 days -20:56:32.643034'),
            (( SELECT id FROM users WHERE email='jonas.weber@example.com' ), '9500', NULL, NULL, NULL, 'Einzahlung', 'deposit', ( SELECT id FROM users WHERE email='bark4-kasse@hadiko.de' ), CURRENT_DATE::timestamptz + INTERVAL '-98 days -08:51:52.267619'),
            (( SELECT id FROM users WHERE email='jonas.weber@example.com' ), '-300', 'Flaschenbier 0,5 L', '150', '2', NULL, 'purchase', ( SELECT id FROM users WHERE email='jonas.weber@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-2348 days -20:02:38.000511'),
            (( SELECT id FROM users WHERE email='jonas.weber@example.com' ), '-540', 'Fassbier Helles', '180', '3', NULL, 'purchase', ( SELECT id FROM users WHERE email='jonas.weber@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-803 days -02:14:40.25076'),
            (( SELECT id FROM users WHERE email='jonas.weber@example.com' ), '-540', 'Fassbier Helles', '180', '3', NULL, 'purchase', ( SELECT id FROM users WHERE email='jonas.weber@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-2339 days -00:23:01.687922'),
            (( SELECT id FROM users WHERE email='jonas.weber@example.com' ), '-150', 'Flaschenbier 0,5 L', '150', '1', NULL, 'purchase', ( SELECT id FROM users WHERE email='jonas.weber@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-1955 days -17:27:13.768295'),
            (( SELECT id FROM users WHERE email='jonas.weber@example.com' ), '-300', 'Flaschenbier 0,5 L', '150', '2', NULL, 'purchase', ( SELECT id FROM users WHERE email='jonas.weber@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-2162 days -06:13:48.122168'),
            (( SELECT id FROM users WHERE email='jonas.weber@example.com' ), '-450', 'Flaschenbier 0,5 L', '150', '3', NULL, 'purchase', ( SELECT id FROM users WHERE email='jonas.weber@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-2268 days -14:15:43.48215'),
            (( SELECT id FROM users WHERE email='jonas.weber@example.com' ), '-300', 'Flaschenbier 0,5 L', '150', '2', NULL, 'purchase', ( SELECT id FROM users WHERE email='jonas.weber@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-1795 days -06:19:56.33867'),
            (( SELECT id FROM users WHERE email='jonas.weber@example.com' ), '-300', 'Flaschenbier 0,5 L', '150', '2', NULL, 'purchase', ( SELECT id FROM users WHERE email='jonas.weber@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-1670 days -22:30:24.441471'),
            (( SELECT id FROM users WHERE email='jonas.weber@example.com' ), '-110', 'Flaschenbier 0,33 L', '110', '1', NULL, 'purchase', ( SELECT id FROM users WHERE email='jonas.weber@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-758 days -13:45:10.935773'),
            (( SELECT id FROM users WHERE email='jonas.weber@example.com' ), '11000', NULL, NULL, NULL, 'Einzahlung', 'deposit', ( SELECT id FROM users WHERE email='bark4-kasse@hadiko.de' ), CURRENT_DATE::timestamptz + INTERVAL '-918 days -11:01:55.666634'),
            (( SELECT id FROM users WHERE email='jonas.weber@example.com' ), '5500', NULL, NULL, NULL, 'Einzahlung', 'deposit', ( SELECT id FROM users WHERE email='bark4-kasse@hadiko.de' ), CURRENT_DATE::timestamptz + INTERVAL '-1554 days -00:30:23.19039'),
            (( SELECT id FROM users WHERE email='jonas.weber@example.com' ), '-300', 'Flaschenbier 0,5 L', '150', '2', NULL, 'purchase', ( SELECT id FROM users WHERE email='jonas.weber@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-1977 days -17:24:53.724259'),
            (( SELECT id FROM users WHERE email='jonas.weber@example.com' ), '-360', 'Fassbier Helles', '180', '2', NULL, 'purchase', ( SELECT id FROM users WHERE email='jonas.weber@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-1684 days -03:44:49.303124'),
            (( SELECT id FROM users WHERE email='jonas.weber@example.com' ), '-360', 'Fassbier Helles', '180', '2', NULL, 'purchase', ( SELECT id FROM users WHERE email='jonas.weber@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-2407 days -18:45:35.028426'),
            (( SELECT id FROM users WHERE email='jonas.weber@example.com' ), '-180', 'Fassbier Helles', '180', '1', NULL, 'purchase', ( SELECT id FROM users WHERE email='jonas.weber@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-1986 days -18:53:07.31017'),
            (( SELECT id FROM users WHERE email='jonas.weber@example.com' ), '-300', 'Flaschenbier 0,5 L', '150', '2', NULL, 'purchase', ( SELECT id FROM users WHERE email='jonas.weber@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-1277 days -11:19:22.217173'),
            (( SELECT id FROM users WHERE email='jonas.weber@example.com' ), '-450', 'Flaschenbier 0,5 L', '150', '3', NULL, 'purchase', ( SELECT id FROM users WHERE email='jonas.weber@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-1458 days -16:47:44.263775'),
            (( SELECT id FROM users WHERE email='jonas.weber@example.com' ), '-180', 'Fassbier Helles', '180', '1', NULL, 'purchase', ( SELECT id FROM users WHERE email='jonas.weber@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-2394 days -12:54:05.040721'),
            (( SELECT id FROM users WHERE email='jonas.weber@example.com' ), '-180', 'Fassbier Helles', '180', '1', NULL, 'purchase', ( SELECT id FROM users WHERE email='jonas.weber@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-1711 days -18:42:06.871839'),
            (( SELECT id FROM users WHERE email='jonas.weber@example.com' ), '-150', 'Flaschenbier 0,5 L', '150', '1', NULL, 'purchase', ( SELECT id FROM users WHERE email='jonas.weber@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-1251 days -18:43:16.319319'),
            (( SELECT id FROM users WHERE email='jonas.weber@example.com' ), '-360', 'Fassbier Helles', '180', '2', NULL, 'purchase', ( SELECT id FROM users WHERE email='jonas.weber@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-1730 days -04:12:15.871181'),
            (( SELECT id FROM users WHERE email='jonas.weber@example.com' ), '3500', NULL, NULL, NULL, 'Einzahlung', 'deposit', ( SELECT id FROM users WHERE email='bark4-kasse@hadiko.de' ), CURRENT_DATE::timestamptz + INTERVAL '-1038 days -22:57:33.835274');
    END IF;
END $$;

-- Transactions for Lea Schmidt (lea.schmidt@example.com) -- 40 rows
DO $$
DECLARE
    uid BIGINT;
    already_seeded BOOLEAN;
BEGIN
    SELECT id INTO uid FROM users WHERE email = 'lea.schmidt@example.com';
    SELECT EXISTS(SELECT 1 FROM transactions WHERE user_id = uid) INTO already_seeded;
    IF NOT already_seeded THEN
        INSERT INTO transactions
            (user_id, amount, item_title, unit_price, quantity, description, type, created_by_user_id, created_at)
        VALUES
            (( SELECT id FROM users WHERE email='lea.schmidt@example.com' ), '-300', 'Flaschenbier 0,5 L', '150', '2', NULL, 'purchase', ( SELECT id FROM users WHERE email='lea.schmidt@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-559 days -23:59:56.148751'),
            (( SELECT id FROM users WHERE email='lea.schmidt@example.com' ), '10000', NULL, NULL, NULL, 'Einzahlung', 'deposit', ( SELECT id FROM users WHERE email='bark4-kasse@hadiko.de' ), CURRENT_DATE::timestamptz + INTERVAL '-1288 days -07:35:11.378434'),
            (( SELECT id FROM users WHERE email='lea.schmidt@example.com' ), '11500', NULL, NULL, NULL, 'Einzahlung', 'deposit', ( SELECT id FROM users WHERE email='bark4-kasse@hadiko.de' ), CURRENT_DATE::timestamptz + INTERVAL '-1172 days -02:57:14.931927'),
            (( SELECT id FROM users WHERE email='lea.schmidt@example.com' ), '9500', NULL, NULL, NULL, 'Einzahlung', 'deposit', ( SELECT id FROM users WHERE email='bark4-kasse@hadiko.de' ), CURRENT_DATE::timestamptz + INTERVAL '-669 days -06:17:06.096677'),
            (( SELECT id FROM users WHERE email='lea.schmidt@example.com' ), '3000', NULL, NULL, NULL, 'Einzahlung', 'deposit', ( SELECT id FROM users WHERE email='bark4-kasse@hadiko.de' ), CURRENT_DATE::timestamptz + INTERVAL '-1042 days -20:43:54.938306'),
            (( SELECT id FROM users WHERE email='lea.schmidt@example.com' ), '-150', 'Flaschenbier 0,5 L', '150', '1', NULL, 'purchase', ( SELECT id FROM users WHERE email='lea.schmidt@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-445 days -01:00:19.81647'),
            (( SELECT id FROM users WHERE email='lea.schmidt@example.com' ), '-180', 'Fassbier Helles', '180', '1', NULL, 'purchase', ( SELECT id FROM users WHERE email='lea.schmidt@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-1172 days -10:55:31.449829'),
            (( SELECT id FROM users WHERE email='lea.schmidt@example.com' ), '-360', 'Fassbier Helles', '180', '2', NULL, 'purchase', ( SELECT id FROM users WHERE email='lea.schmidt@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-415 days -03:47:53.490945'),
            (( SELECT id FROM users WHERE email='lea.schmidt@example.com' ), '-450', 'Flaschenbier 0,5 L', '150', '3', NULL, 'purchase', ( SELECT id FROM users WHERE email='lea.schmidt@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-192 days -09:21:36.510099'),
            (( SELECT id FROM users WHERE email='lea.schmidt@example.com' ), '-540', 'Fassbier Helles', '180', '3', NULL, 'purchase', ( SELECT id FROM users WHERE email='lea.schmidt@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-1379 days -13:51:50.943166'),
            (( SELECT id FROM users WHERE email='lea.schmidt@example.com' ), '-540', 'Fassbier Helles', '180', '3', NULL, 'purchase', ( SELECT id FROM users WHERE email='lea.schmidt@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-765 days -05:41:18.134888'),
            (( SELECT id FROM users WHERE email='lea.schmidt@example.com' ), '-330', 'Flaschenbier 0,33 L', '110', '3', NULL, 'purchase', ( SELECT id FROM users WHERE email='lea.schmidt@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-391 days -10:18:11.213927'),
            (( SELECT id FROM users WHERE email='lea.schmidt@example.com' ), '10500', NULL, NULL, NULL, 'Einzahlung', 'deposit', ( SELECT id FROM users WHERE email='bark4-kasse@hadiko.de' ), CURRENT_DATE::timestamptz + INTERVAL '-210 days -13:18:54.471811'),
            (( SELECT id FROM users WHERE email='lea.schmidt@example.com' ), '-220', 'Flaschenbier 0,33 L', '110', '2', NULL, 'purchase', ( SELECT id FROM users WHERE email='lea.schmidt@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-656 days -14:00:16.834256'),
            (( SELECT id FROM users WHERE email='lea.schmidt@example.com' ), '3000', NULL, NULL, NULL, 'Einzahlung', 'deposit', ( SELECT id FROM users WHERE email='bark4-kasse@hadiko.de' ), CURRENT_DATE::timestamptz + INTERVAL '-1135 days -14:36:19.006862'),
            (( SELECT id FROM users WHERE email='lea.schmidt@example.com' ), '-150', 'Flaschenbier 0,5 L', '150', '1', NULL, 'purchase', ( SELECT id FROM users WHERE email='lea.schmidt@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-312 days -21:24:40.99398'),
            (( SELECT id FROM users WHERE email='lea.schmidt@example.com' ), '-450', 'Flaschenbier 0,5 L', '150', '3', NULL, 'purchase', ( SELECT id FROM users WHERE email='lea.schmidt@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-18 days -11:05:33.105758'),
            (( SELECT id FROM users WHERE email='lea.schmidt@example.com' ), '-450', 'Flaschenbier 0,5 L', '150', '3', NULL, 'purchase', ( SELECT id FROM users WHERE email='lea.schmidt@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-687 days -22:26:49.266018'),
            (( SELECT id FROM users WHERE email='lea.schmidt@example.com' ), '-300', 'Flaschenbier 0,5 L', '150', '2', NULL, 'purchase', ( SELECT id FROM users WHERE email='lea.schmidt@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-121 days -08:01:45.239628'),
            (( SELECT id FROM users WHERE email='lea.schmidt@example.com' ), '-110', 'Flaschenbier 0,33 L', '110', '1', NULL, 'purchase', ( SELECT id FROM users WHERE email='lea.schmidt@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-642 days -09:01:56.662896'),
            (( SELECT id FROM users WHERE email='lea.schmidt@example.com' ), '10000', NULL, NULL, NULL, 'Einzahlung', 'deposit', ( SELECT id FROM users WHERE email='bark4-kasse@hadiko.de' ), CURRENT_DATE::timestamptz + INTERVAL '-644 days -05:55:35.698934'),
            (( SELECT id FROM users WHERE email='lea.schmidt@example.com' ), '-360', 'Fassbier Helles', '180', '2', NULL, 'purchase', ( SELECT id FROM users WHERE email='lea.schmidt@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-1469 days -17:50:27.261753'),
            (( SELECT id FROM users WHERE email='lea.schmidt@example.com' ), '-220', 'Flaschenbier 0,33 L', '110', '2', NULL, 'purchase', ( SELECT id FROM users WHERE email='lea.schmidt@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-1452 days -12:27:08.541812'),
            (( SELECT id FROM users WHERE email='lea.schmidt@example.com' ), '908', NULL, NULL, NULL, 'Sonderbuchung', 'custom', ( SELECT id FROM users WHERE email='bark4-kasse@hadiko.de' ), CURRENT_DATE::timestamptz + INTERVAL '-185 days -23:10:21.828373'),
            (( SELECT id FROM users WHERE email='lea.schmidt@example.com' ), '3000', NULL, NULL, NULL, 'Einzahlung', 'deposit', ( SELECT id FROM users WHERE email='bark4-kasse@hadiko.de' ), CURRENT_DATE::timestamptz + INTERVAL '-75 days -02:06:05.012568'),
            (( SELECT id FROM users WHERE email='lea.schmidt@example.com' ), '2500', NULL, NULL, NULL, 'Einzahlung', 'deposit', ( SELECT id FROM users WHERE email='bark4-kasse@hadiko.de' ), CURRENT_DATE::timestamptz + INTERVAL '-85 days -13:46:49.278524'),
            (( SELECT id FROM users WHERE email='lea.schmidt@example.com' ), '-220', 'Flaschenbier 0,33 L', '110', '2', NULL, 'purchase', ( SELECT id FROM users WHERE email='lea.schmidt@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-1290 days -08:34:41.405763'),
            (( SELECT id FROM users WHERE email='lea.schmidt@example.com' ), '-150', 'Flaschenbier 0,5 L', '150', '1', NULL, 'purchase', ( SELECT id FROM users WHERE email='lea.schmidt@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-1537 days -15:31:04.820382'),
            (( SELECT id FROM users WHERE email='lea.schmidt@example.com' ), '7000', NULL, NULL, NULL, 'Einzahlung', 'deposit', ( SELECT id FROM users WHERE email='bark4-kasse@hadiko.de' ), CURRENT_DATE::timestamptz + INTERVAL '-1198 days -10:37:05.590764'),
            (( SELECT id FROM users WHERE email='lea.schmidt@example.com' ), '-180', 'Fassbier Helles', '180', '1', NULL, 'purchase', ( SELECT id FROM users WHERE email='lea.schmidt@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-514 days -15:34:55.945968'),
            (( SELECT id FROM users WHERE email='lea.schmidt@example.com' ), '-330', 'Flaschenbier 0,33 L', '110', '3', NULL, 'purchase', ( SELECT id FROM users WHERE email='lea.schmidt@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-1587 days -21:10:38.040548'),
            (( SELECT id FROM users WHERE email='lea.schmidt@example.com' ), '-154', NULL, NULL, NULL, 'Sonderbuchung', 'custom', ( SELECT id FROM users WHERE email='bark4-kasse@hadiko.de' ), CURRENT_DATE::timestamptz + INTERVAL '-1044 days -11:03:52.604671'),
            (( SELECT id FROM users WHERE email='lea.schmidt@example.com' ), '-215', NULL, NULL, NULL, 'Sonderbuchung', 'custom', ( SELECT id FROM users WHERE email='bark4-kasse@hadiko.de' ), CURRENT_DATE::timestamptz + INTERVAL '-1384 days -05:40:51.219035'),
            (( SELECT id FROM users WHERE email='lea.schmidt@example.com' ), '-360', 'Fassbier Helles', '180', '2', NULL, 'purchase', ( SELECT id FROM users WHERE email='lea.schmidt@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-580 days -07:33:00.172967'),
            (( SELECT id FROM users WHERE email='lea.schmidt@example.com' ), '7500', NULL, NULL, NULL, 'Einzahlung', 'deposit', ( SELECT id FROM users WHERE email='bark4-kasse@hadiko.de' ), CURRENT_DATE::timestamptz + INTERVAL '-1373 days -16:06:42.685071'),
            (( SELECT id FROM users WHERE email='lea.schmidt@example.com' ), '-540', 'Fassbier Helles', '180', '3', NULL, 'purchase', ( SELECT id FROM users WHERE email='lea.schmidt@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-109 days -13:38:18.446212'),
            (( SELECT id FROM users WHERE email='lea.schmidt@example.com' ), '-540', 'Fassbier Helles', '180', '3', NULL, 'purchase', ( SELECT id FROM users WHERE email='lea.schmidt@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-640 days -14:21:12.703729'),
            (( SELECT id FROM users WHERE email='lea.schmidt@example.com' ), '-150', 'Flaschenbier 0,5 L', '150', '1', NULL, 'purchase', ( SELECT id FROM users WHERE email='lea.schmidt@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-1083 days -00:14:44.457908'),
            (( SELECT id FROM users WHERE email='lea.schmidt@example.com' ), '-220', 'Flaschenbier 0,33 L', '110', '2', NULL, 'purchase', ( SELECT id FROM users WHERE email='lea.schmidt@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-913 days -17:52:48.884124'),
            (( SELECT id FROM users WHERE email='lea.schmidt@example.com' ), '-360', 'Fassbier Helles', '180', '2', NULL, 'purchase', ( SELECT id FROM users WHERE email='lea.schmidt@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-1212 days -11:37:39.2296');
    END IF;
END $$;

-- Transactions for Tim Hoffmann (tim.hoffmann@example.com) -- 30 rows
DO $$
DECLARE
    uid BIGINT;
    already_seeded BOOLEAN;
BEGIN
    SELECT id INTO uid FROM users WHERE email = 'tim.hoffmann@example.com';
    SELECT EXISTS(SELECT 1 FROM transactions WHERE user_id = uid) INTO already_seeded;
    IF NOT already_seeded THEN
        INSERT INTO transactions
            (user_id, amount, item_title, unit_price, quantity, description, type, created_by_user_id, created_at)
        VALUES
            (( SELECT id FROM users WHERE email='tim.hoffmann@example.com' ), '-220', 'Flaschenbier 0,33 L', '110', '2', NULL, 'purchase', ( SELECT id FROM users WHERE email='tim.hoffmann@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-737 days -14:55:23.921409'),
            (( SELECT id FROM users WHERE email='tim.hoffmann@example.com' ), '-632', NULL, NULL, NULL, 'Sonderbuchung', 'custom', ( SELECT id FROM users WHERE email='bark4-kasse@hadiko.de' ), CURRENT_DATE::timestamptz + INTERVAL '-726 days -13:10:02.659088'),
            (( SELECT id FROM users WHERE email='tim.hoffmann@example.com' ), '-675', NULL, NULL, NULL, 'Sonderbuchung', 'custom', ( SELECT id FROM users WHERE email='bark4-kasse@hadiko.de' ), CURRENT_DATE::timestamptz + INTERVAL '-1080 days -03:43:01.030768'),
            (( SELECT id FROM users WHERE email='tim.hoffmann@example.com' ), '3000', NULL, NULL, NULL, 'Einzahlung', 'deposit', ( SELECT id FROM users WHERE email='bark4-kasse@hadiko.de' ), CURRENT_DATE::timestamptz + INTERVAL '-901 days -19:44:32.814618'),
            (( SELECT id FROM users WHERE email='tim.hoffmann@example.com' ), '-220', 'Flaschenbier 0,33 L', '110', '2', NULL, 'purchase', ( SELECT id FROM users WHERE email='tim.hoffmann@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-657 days -11:07:07.257252'),
            (( SELECT id FROM users WHERE email='tim.hoffmann@example.com' ), '-330', 'Flaschenbier 0,33 L', '110', '3', NULL, 'purchase', ( SELECT id FROM users WHERE email='tim.hoffmann@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-778 days -05:32:37.4689'),
            (( SELECT id FROM users WHERE email='tim.hoffmann@example.com' ), '-300', 'Flaschenbier 0,5 L', '150', '2', NULL, 'purchase', ( SELECT id FROM users WHERE email='tim.hoffmann@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-589 days -11:22:04.123121'),
            (( SELECT id FROM users WHERE email='tim.hoffmann@example.com' ), '-110', 'Flaschenbier 0,33 L', '110', '1', NULL, 'purchase', ( SELECT id FROM users WHERE email='tim.hoffmann@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-9 days -04:52:36.080456'),
            (( SELECT id FROM users WHERE email='tim.hoffmann@example.com' ), '-300', 'Flaschenbier 0,5 L', '150', '2', NULL, 'purchase', ( SELECT id FROM users WHERE email='tim.hoffmann@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-441 days -18:48:12.745491'),
            (( SELECT id FROM users WHERE email='tim.hoffmann@example.com' ), '-220', 'Flaschenbier 0,33 L', '110', '2', NULL, 'purchase', ( SELECT id FROM users WHERE email='tim.hoffmann@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-316 days -00:09:23.031468'),
            (( SELECT id FROM users WHERE email='tim.hoffmann@example.com' ), '-220', 'Flaschenbier 0,33 L', '110', '2', NULL, 'purchase', ( SELECT id FROM users WHERE email='tim.hoffmann@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-795 days -14:21:42.768041'),
            (( SELECT id FROM users WHERE email='tim.hoffmann@example.com' ), '11500', NULL, NULL, NULL, 'Einzahlung', 'deposit', ( SELECT id FROM users WHERE email='bark4-kasse@hadiko.de' ), CURRENT_DATE::timestamptz + INTERVAL '-509 days -09:10:11.228934'),
            (( SELECT id FROM users WHERE email='tim.hoffmann@example.com' ), '-220', 'Flaschenbier 0,33 L', '110', '2', NULL, 'purchase', ( SELECT id FROM users WHERE email='tim.hoffmann@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-950 days -16:11:52.069472'),
            (( SELECT id FROM users WHERE email='tim.hoffmann@example.com' ), '-180', 'Fassbier Helles', '180', '1', NULL, 'purchase', ( SELECT id FROM users WHERE email='tim.hoffmann@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-702 days -03:54:05.136381'),
            (( SELECT id FROM users WHERE email='tim.hoffmann@example.com' ), '12000', NULL, NULL, NULL, 'Einzahlung', 'deposit', ( SELECT id FROM users WHERE email='bark4-kasse@hadiko.de' ), CURRENT_DATE::timestamptz + INTERVAL '-180 days -04:25:35.442677'),
            (( SELECT id FROM users WHERE email='tim.hoffmann@example.com' ), '-110', 'Flaschenbier 0,33 L', '110', '1', NULL, 'purchase', ( SELECT id FROM users WHERE email='tim.hoffmann@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-448 days -00:39:27.993053'),
            (( SELECT id FROM users WHERE email='tim.hoffmann@example.com' ), '-150', 'Flaschenbier 0,5 L', '150', '1', NULL, 'purchase', ( SELECT id FROM users WHERE email='tim.hoffmann@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-46 days -22:24:09.126268'),
            (( SELECT id FROM users WHERE email='tim.hoffmann@example.com' ), '5000', NULL, NULL, NULL, 'Einzahlung', 'deposit', ( SELECT id FROM users WHERE email='bark4-kasse@hadiko.de' ), CURRENT_DATE::timestamptz + INTERVAL '-35 days -13:37:04.155495'),
            (( SELECT id FROM users WHERE email='tim.hoffmann@example.com' ), '-110', 'Flaschenbier 0,33 L', '110', '1', NULL, 'purchase', ( SELECT id FROM users WHERE email='tim.hoffmann@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-900 days -12:24:02.450305'),
            (( SELECT id FROM users WHERE email='tim.hoffmann@example.com' ), '7000', NULL, NULL, NULL, 'Einzahlung', 'deposit', ( SELECT id FROM users WHERE email='bark4-kasse@hadiko.de' ), CURRENT_DATE::timestamptz + INTERVAL '-951 days -06:30:47.991549'),
            (( SELECT id FROM users WHERE email='tim.hoffmann@example.com' ), '4000', NULL, NULL, NULL, 'Einzahlung', 'deposit', ( SELECT id FROM users WHERE email='bark4-kasse@hadiko.de' ), CURRENT_DATE::timestamptz + INTERVAL '-969 days -05:14:15.332229'),
            (( SELECT id FROM users WHERE email='tim.hoffmann@example.com' ), '3000', NULL, NULL, NULL, 'Einzahlung', 'deposit', ( SELECT id FROM users WHERE email='bark4-kasse@hadiko.de' ), CURRENT_DATE::timestamptz + INTERVAL '-382 days -07:02:08.714214'),
            (( SELECT id FROM users WHERE email='tim.hoffmann@example.com' ), '-540', NULL, NULL, NULL, 'Sonderbuchung', 'custom', ( SELECT id FROM users WHERE email='bark4-kasse@hadiko.de' ), CURRENT_DATE::timestamptz + INTERVAL '-597 days -17:46:15.529443'),
            (( SELECT id FROM users WHERE email='tim.hoffmann@example.com' ), '11000', NULL, NULL, NULL, 'Einzahlung', 'deposit', ( SELECT id FROM users WHERE email='bark4-kasse@hadiko.de' ), CURRENT_DATE::timestamptz + INTERVAL '-807 days -07:48:30.075929'),
            (( SELECT id FROM users WHERE email='tim.hoffmann@example.com' ), '-220', 'Flaschenbier 0,33 L', '110', '2', NULL, 'purchase', ( SELECT id FROM users WHERE email='tim.hoffmann@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-31 days -03:42:45.061734'),
            (( SELECT id FROM users WHERE email='tim.hoffmann@example.com' ), '278', NULL, NULL, NULL, 'Sonderbuchung', 'custom', ( SELECT id FROM users WHERE email='bark4-kasse@hadiko.de' ), CURRENT_DATE::timestamptz + INTERVAL '-72 days -14:12:15.187895'),
            (( SELECT id FROM users WHERE email='tim.hoffmann@example.com' ), '-220', 'Flaschenbier 0,33 L', '110', '2', NULL, 'purchase', ( SELECT id FROM users WHERE email='tim.hoffmann@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-72 days -04:44:15.520725'),
            (( SELECT id FROM users WHERE email='tim.hoffmann@example.com' ), '587', NULL, NULL, NULL, 'Sonderbuchung', 'custom', ( SELECT id FROM users WHERE email='bark4-kasse@hadiko.de' ), CURRENT_DATE::timestamptz + INTERVAL '-808 days -20:21:11.12525'),
            (( SELECT id FROM users WHERE email='tim.hoffmann@example.com' ), '-450', 'Flaschenbier 0,5 L', '150', '3', NULL, 'purchase', ( SELECT id FROM users WHERE email='tim.hoffmann@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-583 days -00:09:37.651946'),
            (( SELECT id FROM users WHERE email='tim.hoffmann@example.com' ), '-446', NULL, NULL, NULL, 'Sonderbuchung', 'custom', ( SELECT id FROM users WHERE email='bark4-kasse@hadiko.de' ), CURRENT_DATE::timestamptz + INTERVAL '-421 days -13:03:13.335354');
    END IF;
END $$;

-- Transactions for Nina Krueger (nina.krueger@example.com) -- 26 rows
DO $$
DECLARE
    uid BIGINT;
    already_seeded BOOLEAN;
BEGIN
    SELECT id INTO uid FROM users WHERE email = 'nina.krueger@example.com';
    SELECT EXISTS(SELECT 1 FROM transactions WHERE user_id = uid) INTO already_seeded;
    IF NOT already_seeded THEN
        INSERT INTO transactions
            (user_id, amount, item_title, unit_price, quantity, description, type, created_by_user_id, created_at)
        VALUES
            (( SELECT id FROM users WHERE email='nina.krueger@example.com' ), '503', NULL, NULL, NULL, 'Sonderbuchung', 'custom', ( SELECT id FROM users WHERE email='bark4-kasse@hadiko.de' ), CURRENT_DATE::timestamptz + INTERVAL '-1303 days -00:03:56.214663'),
            (( SELECT id FROM users WHERE email='nina.krueger@example.com' ), '-450', 'Flaschenbier 0,5 L', '150', '3', NULL, 'purchase', ( SELECT id FROM users WHERE email='nina.krueger@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-1690 days -04:46:52.818093'),
            (( SELECT id FROM users WHERE email='nina.krueger@example.com' ), '4500', NULL, NULL, NULL, 'Einzahlung', 'deposit', ( SELECT id FROM users WHERE email='bark4-kasse@hadiko.de' ), CURRENT_DATE::timestamptz + INTERVAL '-1363 days -03:25:48.222005'),
            (( SELECT id FROM users WHERE email='nina.krueger@example.com' ), '472', NULL, NULL, NULL, 'Sonderbuchung', 'custom', ( SELECT id FROM users WHERE email='bark4-kasse@hadiko.de' ), CURRENT_DATE::timestamptz + INTERVAL '-1393 days -21:47:40.567148'),
            (( SELECT id FROM users WHERE email='nina.krueger@example.com' ), '-180', 'Fassbier Helles', '180', '1', NULL, 'purchase', ( SELECT id FROM users WHERE email='nina.krueger@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-1217 days -02:57:18.759809'),
            (( SELECT id FROM users WHERE email='nina.krueger@example.com' ), '-300', 'Flaschenbier 0,5 L', '150', '2', NULL, 'purchase', ( SELECT id FROM users WHERE email='nina.krueger@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-2084 days -02:55:23.955932'),
            (( SELECT id FROM users WHERE email='nina.krueger@example.com' ), '-220', 'Flaschenbier 0,33 L', '110', '2', NULL, 'purchase', ( SELECT id FROM users WHERE email='nina.krueger@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-1635 days -03:26:44.176069'),
            (( SELECT id FROM users WHERE email='nina.krueger@example.com' ), '9000', NULL, NULL, NULL, 'Einzahlung', 'deposit', ( SELECT id FROM users WHERE email='bark4-kasse@hadiko.de' ), CURRENT_DATE::timestamptz + INTERVAL '-1475 days -22:01:32.021026'),
            (( SELECT id FROM users WHERE email='nina.krueger@example.com' ), '-450', 'Flaschenbier 0,5 L', '150', '3', NULL, 'purchase', ( SELECT id FROM users WHERE email='nina.krueger@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-1398 days -21:37:08.866949'),
            (( SELECT id FROM users WHERE email='nina.krueger@example.com' ), '9500', NULL, NULL, NULL, 'Einzahlung', 'deposit', ( SELECT id FROM users WHERE email='bark4-kasse@hadiko.de' ), CURRENT_DATE::timestamptz + INTERVAL '-1602 days -17:06:20.80497'),
            (( SELECT id FROM users WHERE email='nina.krueger@example.com' ), '-450', 'Flaschenbier 0,5 L', '150', '3', NULL, 'purchase', ( SELECT id FROM users WHERE email='nina.krueger@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-1499 days -13:57:39.211531'),
            (( SELECT id FROM users WHERE email='nina.krueger@example.com' ), '-220', 'Flaschenbier 0,33 L', '110', '2', NULL, 'purchase', ( SELECT id FROM users WHERE email='nina.krueger@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-2103 days -19:27:55.283282'),
            (( SELECT id FROM users WHERE email='nina.krueger@example.com' ), '-330', 'Flaschenbier 0,33 L', '110', '3', NULL, 'purchase', ( SELECT id FROM users WHERE email='nina.krueger@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-2123 days -01:38:52.055813'),
            (( SELECT id FROM users WHERE email='nina.krueger@example.com' ), '7500', NULL, NULL, NULL, 'Einzahlung', 'deposit', ( SELECT id FROM users WHERE email='bark4-kasse@hadiko.de' ), CURRENT_DATE::timestamptz + INTERVAL '-1870 days -03:07:24.503597'),
            (( SELECT id FROM users WHERE email='nina.krueger@example.com' ), '-360', 'Fassbier Helles', '180', '2', NULL, 'purchase', ( SELECT id FROM users WHERE email='nina.krueger@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-1623 days -20:53:20.77715'),
            (( SELECT id FROM users WHERE email='nina.krueger@example.com' ), '-150', 'Flaschenbier 0,5 L', '150', '1', NULL, 'purchase', ( SELECT id FROM users WHERE email='nina.krueger@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-1261 days -23:08:24.779972'),
            (( SELECT id FROM users WHERE email='nina.krueger@example.com' ), '-450', 'Flaschenbier 0,5 L', '150', '3', NULL, 'purchase', ( SELECT id FROM users WHERE email='nina.krueger@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-2028 days -03:21:37.92841'),
            (( SELECT id FROM users WHERE email='nina.krueger@example.com' ), '-150', 'Flaschenbier 0,5 L', '150', '1', NULL, 'purchase', ( SELECT id FROM users WHERE email='nina.krueger@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-2259 days -20:16:41.8976'),
            (( SELECT id FROM users WHERE email='nina.krueger@example.com' ), '-180', 'Fassbier Helles', '180', '1', NULL, 'purchase', ( SELECT id FROM users WHERE email='nina.krueger@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-1638 days -07:32:13.659202'),
            (( SELECT id FROM users WHERE email='nina.krueger@example.com' ), '-220', 'Flaschenbier 0,33 L', '110', '2', NULL, 'purchase', ( SELECT id FROM users WHERE email='nina.krueger@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-1498 days -18:43:20.93946'),
            (( SELECT id FROM users WHERE email='nina.krueger@example.com' ), '-450', 'Flaschenbier 0,5 L', '150', '3', NULL, 'purchase', ( SELECT id FROM users WHERE email='nina.krueger@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-2153 days -13:05:45.908503'),
            (( SELECT id FROM users WHERE email='nina.krueger@example.com' ), '-217', NULL, NULL, NULL, 'Sonderbuchung', 'custom', ( SELECT id FROM users WHERE email='bark4-kasse@hadiko.de' ), CURRENT_DATE::timestamptz + INTERVAL '-1996 days -20:34:44.140056'),
            (( SELECT id FROM users WHERE email='nina.krueger@example.com' ), '-110', 'Flaschenbier 0,33 L', '110', '1', NULL, 'purchase', ( SELECT id FROM users WHERE email='nina.krueger@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-2076 days -08:57:25.22809'),
            (( SELECT id FROM users WHERE email='nina.krueger@example.com' ), '8500', NULL, NULL, NULL, 'Einzahlung', 'deposit', ( SELECT id FROM users WHERE email='bark4-kasse@hadiko.de' ), CURRENT_DATE::timestamptz + INTERVAL '-1341 days -17:08:52.353344'),
            (( SELECT id FROM users WHERE email='nina.krueger@example.com' ), '-360', 'Fassbier Helles', '180', '2', NULL, 'purchase', ( SELECT id FROM users WHERE email='nina.krueger@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-1936 days -19:39:07.092792'),
            (( SELECT id FROM users WHERE email='nina.krueger@example.com' ), '-34728', NULL, NULL, NULL, 'Guthaben ausgezahlt', 'custom', ( SELECT id FROM users WHERE email='bark4-kasse@hadiko.de' ), CURRENT_DATE::timestamptz + INTERVAL '-1146 days -12:00:00');
    END IF;
END $$;

-- Transactions for Paul Richter (paul.richter@example.com) -- 20 rows
DO $$
DECLARE
    uid BIGINT;
    already_seeded BOOLEAN;
BEGIN
    SELECT id INTO uid FROM users WHERE email = 'paul.richter@example.com';
    SELECT EXISTS(SELECT 1 FROM transactions WHERE user_id = uid) INTO already_seeded;
    IF NOT already_seeded THEN
        INSERT INTO transactions
            (user_id, amount, item_title, unit_price, quantity, description, type, created_by_user_id, created_at)
        VALUES
            (( SELECT id FROM users WHERE email='paul.richter@example.com' ), '7000', NULL, NULL, NULL, 'Einzahlung', 'deposit', ( SELECT id FROM users WHERE email='bark4-kasse@hadiko.de' ), CURRENT_DATE::timestamptz + INTERVAL '-1208 days -12:55:16.826854'),
            (( SELECT id FROM users WHERE email='paul.richter@example.com' ), '-180', 'Fassbier Helles', '180', '1', NULL, 'purchase', ( SELECT id FROM users WHERE email='paul.richter@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-1948 days -04:27:27.371395'),
            (( SELECT id FROM users WHERE email='paul.richter@example.com' ), '-450', 'Flaschenbier 0,5 L', '150', '3', NULL, 'purchase', ( SELECT id FROM users WHERE email='paul.richter@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-1695 days -15:27:31.61381'),
            (( SELECT id FROM users WHERE email='paul.richter@example.com' ), '-330', 'Flaschenbier 0,33 L', '110', '3', NULL, 'purchase', ( SELECT id FROM users WHERE email='paul.richter@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-1678 days -09:59:00.705418'),
            (( SELECT id FROM users WHERE email='paul.richter@example.com' ), '-180', 'Fassbier Helles', '180', '1', NULL, 'purchase', ( SELECT id FROM users WHERE email='paul.richter@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-1767 days -04:06:53.909405'),
            (( SELECT id FROM users WHERE email='paul.richter@example.com' ), '-450', 'Flaschenbier 0,5 L', '150', '3', NULL, 'purchase', ( SELECT id FROM users WHERE email='paul.richter@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-1742 days -19:15:10.974468'),
            (( SELECT id FROM users WHERE email='paul.richter@example.com' ), '-220', 'Flaschenbier 0,33 L', '110', '2', NULL, 'purchase', ( SELECT id FROM users WHERE email='paul.richter@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-1924 days -19:04:16.122625'),
            (( SELECT id FROM users WHERE email='paul.richter@example.com' ), '-540', 'Fassbier Helles', '180', '3', NULL, 'purchase', ( SELECT id FROM users WHERE email='paul.richter@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-1640 days -22:00:40.32035'),
            (( SELECT id FROM users WHERE email='paul.richter@example.com' ), '11000', NULL, NULL, NULL, 'Einzahlung', 'deposit', ( SELECT id FROM users WHERE email='bark4-kasse@hadiko.de' ), CURRENT_DATE::timestamptz + INTERVAL '-1413 days -01:29:19.392788'),
            (( SELECT id FROM users WHERE email='paul.richter@example.com' ), '-450', 'Flaschenbier 0,5 L', '150', '3', NULL, 'purchase', ( SELECT id FROM users WHERE email='paul.richter@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-1996 days -23:33:25.642691'),
            (( SELECT id FROM users WHERE email='paul.richter@example.com' ), '-300', 'Flaschenbier 0,5 L', '150', '2', NULL, 'purchase', ( SELECT id FROM users WHERE email='paul.richter@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-2008 days -09:03:25.996864'),
            (( SELECT id FROM users WHERE email='paul.richter@example.com' ), '5500', NULL, NULL, NULL, 'Einzahlung', 'deposit', ( SELECT id FROM users WHERE email='bark4-kasse@hadiko.de' ), CURRENT_DATE::timestamptz + INTERVAL '-1508 days -06:48:37.754789'),
            (( SELECT id FROM users WHERE email='paul.richter@example.com' ), '-110', 'Flaschenbier 0,33 L', '110', '1', NULL, 'purchase', ( SELECT id FROM users WHERE email='paul.richter@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-1519 days -10:09:46.314447'),
            (( SELECT id FROM users WHERE email='paul.richter@example.com' ), '4500', NULL, NULL, NULL, 'Einzahlung', 'deposit', ( SELECT id FROM users WHERE email='bark4-kasse@hadiko.de' ), CURRENT_DATE::timestamptz + INTERVAL '-1178 days -00:38:22.933471'),
            (( SELECT id FROM users WHERE email='paul.richter@example.com' ), '-150', 'Flaschenbier 0,5 L', '150', '1', NULL, 'purchase', ( SELECT id FROM users WHERE email='paul.richter@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-1688 days -03:10:04.93373'),
            (( SELECT id FROM users WHERE email='paul.richter@example.com' ), '-220', 'Flaschenbier 0,33 L', '110', '2', NULL, 'purchase', ( SELECT id FROM users WHERE email='paul.richter@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-1196 days -21:50:02.698194'),
            (( SELECT id FROM users WHERE email='paul.richter@example.com' ), '-360', 'Fassbier Helles', '180', '2', NULL, 'purchase', ( SELECT id FROM users WHERE email='paul.richter@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-1314 days -19:41:15.914175'),
            (( SELECT id FROM users WHERE email='paul.richter@example.com' ), '-360', 'Fassbier Helles', '180', '2', NULL, 'purchase', ( SELECT id FROM users WHERE email='paul.richter@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-1798 days -10:38:53.489135'),
            (( SELECT id FROM users WHERE email='paul.richter@example.com' ), '-300', 'Flaschenbier 0,5 L', '150', '2', NULL, 'purchase', ( SELECT id FROM users WHERE email='paul.richter@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-1791 days -07:49:02.402434'),
            (( SELECT id FROM users WHERE email='paul.richter@example.com' ), '-220', 'Flaschenbier 0,33 L', '110', '2', NULL, 'purchase', ( SELECT id FROM users WHERE email='paul.richter@example.com' ), CURRENT_DATE::timestamptz + INTERVAL '-1258 days -15:31:04.46177');
    END IF;
END $$;

COMMIT;
