package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/k4-bar/deckel/internal/model"
)

// CreateTransaction inserts a new transaction and returns it with the generated ID and timestamp.
func CreateTransaction(ctx context.Context, db DBTX, t *model.Transaction) (*model.Transaction, error) {
	err := db.QueryRow(ctx, `
		INSERT INTO transactions (user_id, amount, item_title, unit_price, quantity, description, type, cancels_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at`,
		t.UserID, t.Amount, t.ItemTitle, t.UnitPrice, t.Quantity, t.Description, t.Type, t.CancelsID,
	).Scan(&t.ID, &t.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create transaction: %w", err)
	}
	return t, nil
}

// GetTransaction returns the transaction with the given ID, or ErrNotFound if missing.
func GetTransaction(ctx context.Context, db DBTX, id int64) (*model.Transaction, error) {
	var t model.Transaction
	err := db.QueryRow(ctx, `
		SELECT id, user_id, amount, item_title, unit_price, quantity, description,
		       type, cancelled_at, cancels_id, created_at
		FROM transactions WHERE id = $1`, id).Scan(
		&t.ID, &t.UserID, &t.Amount, &t.ItemTitle, &t.UnitPrice, &t.Quantity, &t.Description,
		&t.Type, &t.CancelledAt, &t.CancelsID, &t.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get transaction: %w", err)
	}
	return &t, nil
}

// ListTransactionsByUser returns paginated transactions for a user (newest first) and the total count.
func ListTransactionsByUser(ctx context.Context, db DBTX, userID int64, limit, offset int) ([]model.Transaction, int, error) {
	var total int
	err := db.QueryRow(ctx, `
		SELECT COUNT(*) FROM transactions WHERE user_id = $1`, userID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count transactions by user: %w", err)
	}

	rows, err := db.Query(ctx, `
		SELECT id, user_id, amount, item_title, unit_price, quantity, description,
		       type, cancelled_at, cancels_id, created_at
		FROM transactions
		WHERE user_id = $1
		ORDER BY created_at DESC, id DESC
		LIMIT $2 OFFSET $3`, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list transactions by user: %w", err)
	}
	defer rows.Close()

	var txns []model.Transaction
	for rows.Next() {
		var t model.Transaction
		if err := rows.Scan(
			&t.ID, &t.UserID, &t.Amount, &t.ItemTitle, &t.UnitPrice, &t.Quantity, &t.Description,
			&t.Type, &t.CancelledAt, &t.CancelsID, &t.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan transaction: %w", err)
		}
		txns = append(txns, t)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate transactions: %w", err)
	}

	return txns, total, nil
}

// ListAllTransactionsByUser returns all transactions for a user (newest first), without pagination.
func ListAllTransactionsByUser(ctx context.Context, db DBTX, userID int64) ([]model.Transaction, error) {
	rows, err := db.Query(ctx, `
		SELECT id, user_id, amount, item_title, unit_price, quantity, description,
		       type, cancelled_at, cancels_id, created_at
		FROM transactions
		WHERE user_id = $1
		ORDER BY created_at DESC, id DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("list all transactions by user: %w", err)
	}
	defer rows.Close()

	var txns []model.Transaction
	for rows.Next() {
		var t model.Transaction
		if err := rows.Scan(
			&t.ID, &t.UserID, &t.Amount, &t.ItemTitle, &t.UnitPrice, &t.Quantity, &t.Description,
			&t.Type, &t.CancelledAt, &t.CancelsID, &t.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan transaction: %w", err)
		}
		txns = append(txns, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate transactions: %w", err)
	}

	return txns, nil
}

// ListAllTransactions returns paginated transactions across all users (newest first),
// joined with user info, and the total count.
func ListAllTransactions(ctx context.Context, db DBTX, limit, offset int) ([]model.TransactionWithUser, int, error) {
	var total int
	err := db.QueryRow(ctx, `SELECT COUNT(*) FROM transactions`).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count all transactions: %w", err)
	}

	rows, err := db.Query(ctx, `
		SELECT t.id, t.user_id, t.amount, t.item_title, t.unit_price, t.quantity, t.description,
		       t.type, t.cancelled_at, t.cancels_id, t.created_at,
		       u.full_name, u.email
		FROM transactions t
		JOIN users u ON u.id = t.user_id
		ORDER BY t.created_at DESC, t.id DESC
		LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list all transactions: %w", err)
	}
	defer rows.Close()

	var txns []model.TransactionWithUser
	for rows.Next() {
		var t model.TransactionWithUser
		if err := rows.Scan(
			&t.ID, &t.UserID, &t.Amount, &t.ItemTitle, &t.UnitPrice, &t.Quantity, &t.Description,
			&t.Type, &t.CancelledAt, &t.CancelsID, &t.CreatedAt,
			&t.UserName, &t.UserEmail,
		); err != nil {
			return nil, 0, fmt.Errorf("scan transaction with user: %w", err)
		}
		txns = append(txns, t)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate all transactions: %w", err)
	}

	return txns, total, nil
}

// CountTransactionsByUser returns the total number of transactions for a given user.
func CountTransactionsByUser(ctx context.Context, db DBTX, userID int64) (int, error) {
	var count int
	err := db.QueryRow(ctx, `
		SELECT COUNT(*) FROM transactions WHERE user_id = $1`, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count transactions by user: %w", err)
	}
	return count, nil
}

// CancelTransaction sets cancelled_at on the original transaction and inserts a
// counter-transaction with type=cancellation, amount=-original.amount, and cancels_id pointing
// to the original. Both the original and counter-transaction get cancelled_at = NOW().
func CancelTransaction(ctx context.Context, db DBTX, id int64) error {
	now := time.Now()

	// Set cancelled_at on original transaction.
	tag, err := db.Exec(ctx, `
		UPDATE transactions SET cancelled_at = $2 WHERE id = $1 AND cancelled_at IS NULL`, id, now)
	if err != nil {
		return fmt.Errorf("cancel transaction: update original: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	// Fetch original to get amount, user_id, and description for counter-transaction.
	var userID int64
	var amount int64
	var itemTitle, description *string
	err = db.QueryRow(ctx, `
		SELECT user_id, amount, item_title, description FROM transactions WHERE id = $1`, id).Scan(&userID, &amount, &itemTitle, &description)
	if err != nil {
		return fmt.Errorf("cancel transaction: get original: %w", err)
	}

	// Build storno description from original item title or description.
	stornoDesc := "Storno"
	if itemTitle != nil && *itemTitle != "" {
		stornoDesc = "Storno: " + *itemTitle
	} else if description != nil && *description != "" {
		stornoDesc = "Storno: " + *description
	}

	// Insert counter-transaction (no cancelled_at — it's the active reversal record).
	_, err = db.Exec(ctx, `
		INSERT INTO transactions (user_id, amount, description, type, cancels_id)
		VALUES ($1, $2, $3, 'cancellation', $4)`,
		userID, -amount, stornoDesc, id)
	if err != nil {
		return fmt.Errorf("cancel transaction: insert counter: %w", err)
	}

	return nil
}
