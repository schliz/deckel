package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/k4-bar/deckel/internal/model"
)

// ErrNotFound indicates that the requested record does not exist.
var ErrNotFound = errors.New("not found")

// GetUserByEmail returns the user with the given email, or nil + ErrNotFound if missing.
func GetUserByEmail(ctx context.Context, db DBTX, email string) (*model.User, error) {
	var u model.User
	err := db.QueryRow(ctx, `
		SELECT id, email, full_name, given_name, family_name,
		       is_barteamer, is_admin, is_active, spending_limit_disabled,
		       created_at, updated_at
		FROM users WHERE email = $1`, email).Scan(
		&u.ID, &u.Email, &u.FullName, &u.GivenName, &u.FamilyName,
		&u.IsBarteamer, &u.IsAdmin, &u.IsActive, &u.SpendingLimitDisabled,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return &u, nil
}

// CreateUser inserts a new user and returns it with the generated ID and timestamps.
func CreateUser(ctx context.Context, db DBTX, u *model.User) (*model.User, error) {
	err := db.QueryRow(ctx, `
		INSERT INTO users (email, full_name, given_name, family_name, is_barteamer, is_admin, is_active, spending_limit_disabled)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at`,
		u.Email, u.FullName, u.GivenName, u.FamilyName,
		u.IsBarteamer, u.IsAdmin, u.IsActive, u.SpendingLimitDisabled,
	).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return u, nil
}

// UpdateUserProfile updates a user's profile fields and updated_at timestamp.
func UpdateUserProfile(ctx context.Context, db DBTX, id int64, fullName, givenName, familyName string, isAdmin bool) error {
	_, err := db.Exec(ctx, `
		UPDATE users
		SET full_name = $2, given_name = $3, family_name = $4, is_admin = $5, updated_at = NOW()
		WHERE id = $1`,
		id, fullName, givenName, familyName, isAdmin,
	)
	if err != nil {
		return fmt.Errorf("update user profile: %w", err)
	}
	return nil
}

// GetUserBalance returns the sum of all non-cancelled transaction amounts for the given user.
func GetUserBalance(ctx context.Context, db DBTX, userID int64) (int64, error) {
	var balance int64
	err := db.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount), 0) FROM transactions
		WHERE user_id = $1 AND cancelled_at IS NULL`, userID).Scan(&balance)
	if err != nil {
		return 0, fmt.Errorf("get user balance: %w", err)
	}
	return balance, nil
}

// GetUserRank returns the user's rank (1 = highest balance) and total active user count.
func GetUserRank(ctx context.Context, db DBTX, userID int64) (rank int, total int, err error) {
	err = db.QueryRow(ctx, `
		SELECT COUNT(*) + 1 FROM (
			SELECT u.id, COALESCE(SUM(t.amount), 0) as balance
			FROM users u
			LEFT JOIN transactions t ON t.user_id = u.id AND t.cancelled_at IS NULL
			WHERE u.is_active = TRUE
			GROUP BY u.id
			HAVING COALESCE(SUM(t.amount), 0) > (
				SELECT COALESCE(SUM(amount), 0) FROM transactions
				WHERE user_id = $1 AND cancelled_at IS NULL
			)
		) as higher`, userID).Scan(&rank)
	if err != nil {
		return 0, 0, fmt.Errorf("get user rank: %w", err)
	}

	err = db.QueryRow(ctx, `
		SELECT COUNT(*) FROM users WHERE is_active = TRUE`).Scan(&total)
	if err != nil {
		return 0, 0, fmt.Errorf("get total active users: %w", err)
	}

	return rank, total, nil
}

// GetAllBalancesSum returns the sum of all active users' balances (non-cancelled transactions).
func GetAllBalancesSum(ctx context.Context, db DBTX) (int64, error) {
	var sum int64
	err := db.QueryRow(ctx, `
		SELECT COALESCE(SUM(sub.balance), 0) FROM (
			SELECT COALESCE(SUM(t.amount), 0) as balance
			FROM users u
			LEFT JOIN transactions t ON t.user_id = u.id AND t.cancelled_at IS NULL
			WHERE u.is_active = TRUE
			GROUP BY u.id
		) sub`).Scan(&sum)
	if err != nil {
		return 0, fmt.Errorf("get all balances sum: %w", err)
	}
	return sum, nil
}
