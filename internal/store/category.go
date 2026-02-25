package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/k4-bar/deckel/internal/model"
)

// ListCategories returns all categories ordered by sort_order.
func ListCategories(ctx context.Context, db DBTX) ([]model.Category, error) {
	rows, err := db.Query(ctx, `
		SELECT id, name, sort_order, created_at
		FROM categories ORDER BY sort_order, id`)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	defer rows.Close()

	var cats []model.Category
	for rows.Next() {
		var c model.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.SortOrder, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("list categories scan: %w", err)
		}
		cats = append(cats, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list categories rows: %w", err)
	}
	return cats, nil
}

// GetCategory returns a single category by ID.
func GetCategory(ctx context.Context, db DBTX, id int64) (*model.Category, error) {
	var c model.Category
	err := db.QueryRow(ctx, `
		SELECT id, name, sort_order, created_at
		FROM categories WHERE id = $1`, id).Scan(
		&c.ID, &c.Name, &c.SortOrder, &c.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get category: %w", err)
	}
	return &c, nil
}

// CreateCategory inserts a new category with the given name. The sort_order is
// set to one more than the current maximum so the new category appears last.
func CreateCategory(ctx context.Context, db DBTX, name string) (*model.Category, error) {
	var c model.Category
	err := db.QueryRow(ctx, `
		INSERT INTO categories (name, sort_order)
		VALUES ($1, COALESCE((SELECT MAX(sort_order) FROM categories), 0) + 1)
		RETURNING id, name, sort_order, created_at`, name).Scan(
		&c.ID, &c.Name, &c.SortOrder, &c.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create category: %w", err)
	}
	return &c, nil
}

// UpdateCategory renames a category.
func UpdateCategory(ctx context.Context, db DBTX, id int64, name string) error {
	ct, err := db.Exec(ctx, `UPDATE categories SET name = $1 WHERE id = $2`, name, id)
	if err != nil {
		return fmt.Errorf("update category: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// DeleteCategory removes a category by ID.
func DeleteCategory(ctx context.Context, db DBTX, id int64) error {
	ct, err := db.Exec(ctx, `DELETE FROM categories WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete category: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ReorderCategory changes a category's sort position. Direction must be one of:
// "up" (swap with previous), "down" (swap with next), "top" (move to first), "bottom" (move to last).
func ReorderCategory(ctx context.Context, db DBTX, id int64, direction string) error {
	switch direction {
	case "up":
		return reorderCategorySwap(ctx, db, id, true)
	case "down":
		return reorderCategorySwap(ctx, db, id, false)
	case "top":
		return reorderCategoryEdge(ctx, db, id, true)
	case "bottom":
		return reorderCategoryEdge(ctx, db, id, false)
	default:
		return fmt.Errorf("invalid reorder direction: %s", direction)
	}
}

// reorderCategorySwap swaps the sort_order of the target category with its
// immediate neighbor. If moveUp is true, swaps with the category that has the
// next lower sort_order; otherwise swaps with the next higher.
func reorderCategorySwap(ctx context.Context, db DBTX, id int64, moveUp bool) error {
	// Get current sort_order.
	var currentOrder int
	err := db.QueryRow(ctx, `SELECT sort_order FROM categories WHERE id = $1`, id).Scan(&currentOrder)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("reorder category get current: %w", err)
	}

	// Find the neighbor to swap with.
	var neighborID int64
	var neighborOrder int
	var query string
	if moveUp {
		query = `SELECT id, sort_order FROM categories WHERE sort_order < $1 ORDER BY sort_order DESC LIMIT 1`
	} else {
		query = `SELECT id, sort_order FROM categories WHERE sort_order > $1 ORDER BY sort_order ASC LIMIT 1`
	}
	err = db.QueryRow(ctx, query, currentOrder).Scan(&neighborID, &neighborOrder)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil // already at edge, no-op
		}
		return fmt.Errorf("reorder category find neighbor: %w", err)
	}

	// Swap sort_orders.
	_, err = db.Exec(ctx, `UPDATE categories SET sort_order = $1 WHERE id = $2`, neighborOrder, id)
	if err != nil {
		return fmt.Errorf("reorder category update target: %w", err)
	}
	_, err = db.Exec(ctx, `UPDATE categories SET sort_order = $1 WHERE id = $2`, currentOrder, neighborID)
	if err != nil {
		return fmt.Errorf("reorder category update neighbor: %w", err)
	}
	return nil
}

// reorderCategoryEdge moves a category to the top (sort_order = min - 1) or
// bottom (sort_order = max + 1).
func reorderCategoryEdge(ctx context.Context, db DBTX, id int64, toTop bool) error {
	// Verify category exists.
	var currentOrder int
	err := db.QueryRow(ctx, `SELECT sort_order FROM categories WHERE id = $1`, id).Scan(&currentOrder)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("reorder category get current: %w", err)
	}

	var newOrder int
	if toTop {
		err = db.QueryRow(ctx, `SELECT COALESCE(MIN(sort_order), 0) - 1 FROM categories`).Scan(&newOrder)
	} else {
		err = db.QueryRow(ctx, `SELECT COALESCE(MAX(sort_order), 0) + 1 FROM categories`).Scan(&newOrder)
	}
	if err != nil {
		return fmt.Errorf("reorder category calc edge: %w", err)
	}

	_, err = db.Exec(ctx, `UPDATE categories SET sort_order = $1 WHERE id = $2`, newOrder, id)
	if err != nil {
		return fmt.Errorf("reorder category update: %w", err)
	}
	return nil
}
