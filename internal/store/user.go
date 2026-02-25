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

// GetUserBalanceForUpdate returns the user's balance with a FOR UPDATE lock on the user row.
// This must be called within a transaction to prevent concurrent balance modifications.
func GetUserBalanceForUpdate(ctx context.Context, db DBTX, userID int64) (int64, error) {
	// Lock the user row to prevent concurrent modifications.
	_, err := db.Exec(ctx, `SELECT id FROM users WHERE id = $1 FOR UPDATE`, userID)
	if err != nil {
		return 0, fmt.Errorf("lock user row: %w", err)
	}

	var balance int64
	err = db.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount), 0) FROM transactions
		WHERE user_id = $1`, userID).Scan(&balance)
	if err != nil {
		return 0, fmt.Errorf("get user balance for update: %w", err)
	}
	return balance, nil
}

// GetUserBalance returns the sum of all transaction amounts for the given user.
func GetUserBalance(ctx context.Context, db DBTX, userID int64) (int64, error) {
	var balance int64
	err := db.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount), 0) FROM transactions
		WHERE user_id = $1`, userID).Scan(&balance)
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
			LEFT JOIN transactions t ON t.user_id = u.id
			WHERE u.is_active = TRUE
			GROUP BY u.id
			HAVING COALESCE(SUM(t.amount), 0) > (
				SELECT COALESCE(SUM(amount), 0) FROM transactions
				WHERE user_id = $1
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

// GetAllBalancesSum returns the sum of all active users' balances.
func GetAllBalancesSum(ctx context.Context, db DBTX) (int64, error) {
	var sum int64
	err := db.QueryRow(ctx, `
		SELECT COALESCE(SUM(sub.balance), 0) FROM (
			SELECT COALESCE(SUM(t.amount), 0) as balance
			FROM users u
			LEFT JOIN transactions t ON t.user_id = u.id
			WHERE u.is_active = TRUE
			GROUP BY u.id
		) sub`).Scan(&sum)
	if err != nil {
		return 0, fmt.Errorf("get all balances sum: %w", err)
	}
	return sum, nil
}

// GetNegativeBalancesSum returns the sum of balances for active users who have a negative balance.
func GetNegativeBalancesSum(ctx context.Context, db DBTX) (int64, error) {
	var sum int64
	err := db.QueryRow(ctx, `
		SELECT COALESCE(SUM(sub.balance), 0) FROM (
			SELECT COALESCE(SUM(t.amount), 0) as balance
			FROM users u
			LEFT JOIN transactions t ON t.user_id = u.id
			WHERE u.is_active = TRUE
			GROUP BY u.id
			HAVING COALESCE(SUM(t.amount), 0) < 0
		) sub`).Scan(&sum)
	if err != nil {
		return 0, fmt.Errorf("get negative balances sum: %w", err)
	}
	return sum, nil
}

// ListUsersWithBalance returns a paginated list of users with their computed balance,
// sorted by balance ascending. It also returns the total count of users.
func ListUsersWithBalance(ctx context.Context, db DBTX, limit, offset int) ([]model.UserWithBalance, int, error) {
	var total int
	err := db.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	rows, err := db.Query(ctx, `
		SELECT u.id, u.email, u.full_name, u.given_name, u.family_name,
		       u.is_barteamer, u.is_admin, u.is_active, u.spending_limit_disabled,
		       u.created_at, u.updated_at,
		       COALESCE(SUM(t.amount), 0) AS balance
		FROM users u
		LEFT JOIN transactions t ON t.user_id = u.id
		GROUP BY u.id
		ORDER BY balance ASC
		LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list users with balance: %w", err)
	}
	defer rows.Close()

	var users []model.UserWithBalance
	for rows.Next() {
		var ub model.UserWithBalance
		if err := rows.Scan(
			&ub.ID, &ub.Email, &ub.FullName, &ub.GivenName, &ub.FamilyName,
			&ub.IsBarteamer, &ub.IsAdmin, &ub.IsActive, &ub.SpendingLimitDisabled,
			&ub.CreatedAt, &ub.UpdatedAt,
			&ub.Balance,
		); err != nil {
			return nil, 0, fmt.Errorf("scan user with balance: %w", err)
		}
		users = append(users, ub)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate users with balance: %w", err)
	}

	return users, total, nil
}

// ListActiveUsersWithBalance returns all active users with their computed balance.
func ListActiveUsersWithBalance(ctx context.Context, db DBTX) ([]model.UserWithBalance, error) {
	rows, err := db.Query(ctx, `
		SELECT u.id, u.email, u.full_name, u.given_name, u.family_name,
		       u.is_barteamer, u.is_admin, u.is_active, u.spending_limit_disabled,
		       u.created_at, u.updated_at,
		       COALESCE(SUM(t.amount), 0) AS balance
		FROM users u
		LEFT JOIN transactions t ON t.user_id = u.id
		WHERE u.is_active = TRUE
		GROUP BY u.id
		ORDER BY u.full_name ASC`)
	if err != nil {
		return nil, fmt.Errorf("list active users with balance: %w", err)
	}
	defer rows.Close()

	var users []model.UserWithBalance
	for rows.Next() {
		var ub model.UserWithBalance
		if err := rows.Scan(
			&ub.ID, &ub.Email, &ub.FullName, &ub.GivenName, &ub.FamilyName,
			&ub.IsBarteamer, &ub.IsAdmin, &ub.IsActive, &ub.SpendingLimitDisabled,
			&ub.CreatedAt, &ub.UpdatedAt,
			&ub.Balance,
		); err != nil {
			return nil, fmt.Errorf("scan active user with balance: %w", err)
		}
		users = append(users, ub)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate active users with balance: %w", err)
	}

	return users, nil
}

// ToggleBarteamer flips the is_barteamer flag for the given user.
func ToggleBarteamer(ctx context.Context, db DBTX, id int64) error {
	ct, err := db.Exec(ctx,
		`UPDATE users SET is_barteamer = NOT is_barteamer, updated_at = NOW() WHERE id = $1`,
		id,
	)
	if err != nil {
		return fmt.Errorf("toggle barteamer: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ToggleActive flips the is_active flag for the given user.
func ToggleActive(ctx context.Context, db DBTX, id int64) error {
	ct, err := db.Exec(ctx,
		`UPDATE users SET is_active = NOT is_active, updated_at = NOW() WHERE id = $1`,
		id,
	)
	if err != nil {
		return fmt.Errorf("toggle active: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ToggleSpendingLimit flips the spending_limit_disabled flag for the given user.
func ToggleSpendingLimit(ctx context.Context, db DBTX, id int64) error {
	ct, err := db.Exec(ctx,
		`UPDATE users SET spending_limit_disabled = NOT spending_limit_disabled, updated_at = NOW() WHERE id = $1`,
		id,
	)
	if err != nil {
		return fmt.Errorf("toggle spending limit: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// GetUserWithBalance returns a single user with their computed balance.
func GetUserWithBalance(ctx context.Context, db DBTX, id int64) (*model.UserWithBalance, error) {
	var ub model.UserWithBalance
	err := db.QueryRow(ctx, `
		SELECT u.id, u.email, u.full_name, u.given_name, u.family_name,
		       u.is_barteamer, u.is_admin, u.is_active, u.spending_limit_disabled,
		       u.created_at, u.updated_at,
		       COALESCE(SUM(t.amount), 0) AS balance
		FROM users u
		LEFT JOIN transactions t ON t.user_id = u.id
		WHERE u.id = $1
		GROUP BY u.id`, id).Scan(
		&ub.ID, &ub.Email, &ub.FullName, &ub.GivenName, &ub.FamilyName,
		&ub.IsBarteamer, &ub.IsAdmin, &ub.IsActive, &ub.SpendingLimitDisabled,
		&ub.CreatedAt, &ub.UpdatedAt,
		&ub.Balance,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get user with balance: %w", err)
	}
	return &ub, nil
}
