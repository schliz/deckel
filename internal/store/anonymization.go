package store

import (
	"context"
	"fmt"
	"time"

	"github.com/schliz/deckel/internal/model"
)

// dueCondition is the shared due-date / cancellation-pair predicate used by
// both CountAnonymizableTransactions and AnonymizeTransactions: a
// transaction is due for anonymization when it predates the cutoff and has
// no cancellation counterpart (in either direction of cancels_id) that is
// still on/after the cutoff. Cancellation pairs always move together.
const dueCondition = `
		t.user_id IS NOT NULL AND t.created_at < $1
		AND NOT EXISTS (SELECT 1 FROM transactions p
			WHERE (p.cancels_id = t.id OR t.cancels_id = p.id) AND p.created_at >= $1)`

// CountAnonymizableTransactions returns how many transactions (and how many
// distinct users) are due for anonymization as of the given cutoff.
func CountAnonymizableTransactions(ctx context.Context, db DBTX, cutoff time.Time) (txCount, userCount int, err error) {
	err = db.QueryRow(ctx, `
		SELECT COUNT(*), COUNT(DISTINCT t.user_id)
		FROM transactions t
		WHERE `+dueCondition, cutoff).Scan(&txCount, &userCount)
	if err != nil {
		return 0, 0, fmt.Errorf("count anonymizable transactions: %w", err)
	}
	return txCount, userCount, nil
}

// AnonymizeTransactions anonymizes all transactions due as of cutoff in a
// single atomic statement: a `due` CTE selects the due rows, a `credit` CTE
// adds each affected user's due amounts to users.balance_carryover
// (preserving the balance invariant), and the final UPDATE clears
// user_id/created_by_user_id and stamps anonymized_at on those rows.
// Returns the number of anonymized transactions and the number of distinct
// users affected.
func AnonymizeTransactions(ctx context.Context, db DBTX, cutoff time.Time) (txCount, userCount int, err error) {
	rows, err := db.Query(ctx, `
		WITH due AS (
			SELECT t.id, t.user_id, t.amount
			FROM transactions t
			WHERE `+dueCondition+`
		), credit AS (
			UPDATE users u
			SET balance_carryover = u.balance_carryover + agg.total
			FROM (SELECT user_id, SUM(amount) AS total FROM due GROUP BY user_id) agg
			WHERE u.id = agg.user_id
			RETURNING u.id
		)
		UPDATE transactions t
		SET user_id = NULL, created_by_user_id = NULL, anonymized_at = NOW()
		FROM due
		WHERE t.id = due.id
		RETURNING due.user_id`, cutoff)
	if err != nil {
		return 0, 0, fmt.Errorf("anonymize transactions: %w", err)
	}
	defer rows.Close()

	users := make(map[int64]struct{})
	for rows.Next() {
		var userID int64
		if err := rows.Scan(&userID); err != nil {
			return 0, 0, fmt.Errorf("scan anonymized transaction: %w", err)
		}
		txCount++
		users[userID] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return 0, 0, fmt.Errorf("iterate anonymized transactions: %w", err)
	}

	return txCount, len(users), nil
}

// InsertAnonymizationRun records a completed anonymization run for the GoBD
// audit log, filling in the generated ID and executed_at timestamp.
func InsertAnonymizationRun(ctx context.Context, db DBTX, run *model.AnonymizationRun) error {
	err := db.QueryRow(ctx, `
		INSERT INTO anonymization_runs (cutoff_date, transaction_count, user_count, executed_by)
		VALUES ($1, $2, $3, $4)
		RETURNING id, executed_at`,
		run.CutoffDate, run.TransactionCount, run.UserCount, run.ExecutedBy,
	).Scan(&run.ID, &run.ExecutedAt)
	if err != nil {
		return fmt.Errorf("insert anonymization run: %w", err)
	}
	return nil
}

// ListAnonymizationRuns returns the most recent anonymization runs (newest first), up to limit.
func ListAnonymizationRuns(ctx context.Context, db DBTX, limit int) ([]model.AnonymizationRun, error) {
	rows, err := db.Query(ctx, `
		SELECT id, executed_at, cutoff_date, transaction_count, user_count, executed_by
		FROM anonymization_runs
		ORDER BY executed_at DESC, id DESC
		LIMIT $1`, limit)
	if err != nil {
		return nil, fmt.Errorf("list anonymization runs: %w", err)
	}
	defer rows.Close()

	var runs []model.AnonymizationRun
	for rows.Next() {
		var run model.AnonymizationRun
		if err := rows.Scan(
			&run.ID, &run.ExecutedAt, &run.CutoffDate, &run.TransactionCount, &run.UserCount, &run.ExecutedBy,
		); err != nil {
			return nil, fmt.Errorf("scan anonymization run: %w", err)
		}
		runs = append(runs, run)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate anonymization runs: %w", err)
	}

	return runs, nil
}
