// Package anonymize implements the GDPR anonymization run: it removes the
// person-link (user_id/created_by_user_id) from transactions older than a
// configurable retention period while preserving accounting data (GoBD) and
// the affected users' balances via users.balance_carryover.
//
// The service is deliberately free of HTTP context so a later background
// job (e.g. a goroutine started in cmd/server) can invoke Run with
// executedBy = "system" on the same code path as the admin-triggered run.
package anonymize

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/schliz/deckel/internal/model"
	"github.com/schliz/deckel/internal/store"
)

// ErrDisabled indicates that anonymization is disabled (months <= 0).
var ErrDisabled = errors.New("anonymization is disabled")

// anonymizeLockKey is the pg_advisory_xact_lock key used to serialize
// concurrent anonymization runs, preventing a balance_carryover credit from
// being applied twice if two runs overlap. It is the ASCII bytes of "anon"
// packed into a 32-bit int (0x616e6f6e) — memorable and, since no other
// advisory locks exist in this codebase (verified by search), collision-free.
const anonymizeLockKey int64 = 0x616e6f6e

// Result summarizes a completed (or no-op) anonymization run.
type Result struct {
	Cutoff           time.Time
	TransactionCount int
	UserCount        int
	// Ran is false when nothing was due for anonymization; in that case no
	// anonymization_runs log entry was written.
	Ran bool
}

// Preview reports how many transactions (and distinct users) would be
// anonymized under the given retention period, without changing anything.
// months <= 0 (anonymization disabled) returns ErrDisabled.
func Preview(ctx context.Context, db store.DBTX, months int) (cutoff time.Time, txCount, userCount int, err error) {
	if months <= 0 {
		return time.Time{}, 0, 0, ErrDisabled
	}

	cutoff = time.Now().AddDate(0, -months, 0)
	txCount, userCount, err = store.CountAnonymizableTransactions(ctx, db, cutoff)
	if err != nil {
		return time.Time{}, 0, 0, err
	}
	return cutoff, txCount, userCount, nil
}

// Run executes an anonymization run for transactions older than months,
// crediting each affected user's balance_carryover and logging the run for
// the GoBD audit trail. months <= 0 returns ErrDisabled. If nothing is due,
// no log entry is written and the returned Result has Ran = false.
func Run(ctx context.Context, st *store.Store, months int, executedBy string) (*Result, error) {
	if months <= 0 {
		return nil, ErrDisabled
	}

	cutoff := time.Now().AddDate(0, -months, 0)
	result := &Result{Cutoff: cutoff}

	err := st.WithTx(ctx, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock($1)`, anonymizeLockKey); err != nil {
			return err
		}

		txCount, userCount, err := store.AnonymizeTransactions(ctx, tx, cutoff)
		if err != nil {
			return err
		}
		result.TransactionCount = txCount
		result.UserCount = userCount

		if txCount == 0 {
			return nil
		}
		result.Ran = true

		run := &model.AnonymizationRun{
			CutoffDate:       cutoff,
			TransactionCount: txCount,
			UserCount:        userCount,
			ExecutedBy:       executedBy,
		}
		return store.InsertAnonymizationRun(ctx, tx, run)
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}
