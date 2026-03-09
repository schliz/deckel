package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/schliz/deckel/internal/model"
)

// ListItemsByCategory returns all active (non-soft-deleted) items for the given
// category, ordered by sort_order.
func ListItemsByCategory(ctx context.Context, db DBTX, categoryID int64) ([]model.Item, error) {
	rows, err := db.Query(ctx, `
		SELECT id, category_id, name, price_barteamer, price_helfer, sort_order, deleted_at, created_at, updated_at
		FROM items
		WHERE category_id = $1 AND deleted_at IS NULL
		ORDER BY sort_order, id`, categoryID)
	if err != nil {
		return nil, fmt.Errorf("list items by category: %w", err)
	}
	defer rows.Close()

	var items []model.Item
	for rows.Next() {
		var it model.Item
		if err := rows.Scan(&it.ID, &it.CategoryID, &it.Name, &it.PriceBarteamer, &it.PriceHelfer, &it.SortOrder, &it.DeletedAt, &it.CreatedAt, &it.UpdatedAt); err != nil {
			return nil, fmt.Errorf("list items scan: %w", err)
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list items rows: %w", err)
	}
	return items, nil
}

// GetItem returns an item by ID, including soft-deleted items.
func GetItem(ctx context.Context, db DBTX, id int64) (*model.Item, error) {
	var it model.Item
	err := db.QueryRow(ctx, `
		SELECT id, category_id, name, price_barteamer, price_helfer, sort_order, deleted_at, created_at, updated_at
		FROM items WHERE id = $1`, id).Scan(
		&it.ID, &it.CategoryID, &it.Name, &it.PriceBarteamer, &it.PriceHelfer, &it.SortOrder, &it.DeletedAt, &it.CreatedAt, &it.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get item: %w", err)
	}
	return &it, nil
}

// CreateItem inserts a new item. The sort_order is set to one more than the
// current maximum within the same category so the new item appears last.
func CreateItem(ctx context.Context, db DBTX, item *model.Item) (*model.Item, error) {
	var it model.Item
	err := db.QueryRow(ctx, `
		INSERT INTO items (category_id, name, price_barteamer, price_helfer, sort_order)
		VALUES ($1, $2, $3, $4, COALESCE((SELECT MAX(sort_order) FROM items WHERE category_id = $1 AND deleted_at IS NULL), 0) + 1)
		RETURNING id, category_id, name, price_barteamer, price_helfer, sort_order, deleted_at, created_at, updated_at`,
		item.CategoryID, item.Name, item.PriceBarteamer, item.PriceHelfer).Scan(
		&it.ID, &it.CategoryID, &it.Name, &it.PriceBarteamer, &it.PriceHelfer, &it.SortOrder, &it.DeletedAt, &it.CreatedAt, &it.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create item: %w", err)
	}
	return &it, nil
}

// UpdateItem updates an item's name and prices.
func UpdateItem(ctx context.Context, db DBTX, id int64, name string, priceBarteamer, priceHelfer int64) error {
	ct, err := db.Exec(ctx, `
		UPDATE items SET name = $1, price_barteamer = $2, price_helfer = $3, updated_at = NOW()
		WHERE id = $4 AND deleted_at IS NULL`, name, priceBarteamer, priceHelfer, id)
	if err != nil {
		return fmt.Errorf("update item: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// SoftDeleteItem marks an item as deleted by setting deleted_at to NOW().
func SoftDeleteItem(ctx context.Context, db DBTX, id int64) error {
	ct, err := db.Exec(ctx, `
		UPDATE items SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL`, id)
	if err != nil {
		return fmt.Errorf("soft delete item: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ReorderItem changes an item's sort position within its category.
// Direction must be one of: "up", "down", "top", "bottom".
func ReorderItem(ctx context.Context, db DBTX, id int64, direction string) error {
	switch direction {
	case "up":
		return reorderItemSwap(ctx, db, id, true)
	case "down":
		return reorderItemSwap(ctx, db, id, false)
	case "top":
		return reorderItemEdge(ctx, db, id, true)
	case "bottom":
		return reorderItemEdge(ctx, db, id, false)
	default:
		return fmt.Errorf("invalid reorder direction: %s", direction)
	}
}

// CountActiveItemsByCategory returns the number of active (non-soft-deleted)
// items in the given category.
func CountActiveItemsByCategory(ctx context.Context, db DBTX, categoryID int64) (int, error) {
	var count int
	err := db.QueryRow(ctx, `
		SELECT COUNT(*) FROM items
		WHERE category_id = $1 AND deleted_at IS NULL`, categoryID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count active items: %w", err)
	}
	return count, nil
}

func reorderItemSwap(ctx context.Context, db DBTX, id int64, moveUp bool) error {
	// Get current item's sort_order and category_id.
	var currentOrder int
	var categoryID int64
	err := db.QueryRow(ctx, `
		SELECT sort_order, category_id FROM items
		WHERE id = $1 AND deleted_at IS NULL`, id).Scan(&currentOrder, &categoryID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("reorder item get current: %w", err)
	}

	// Find the neighbor to swap with (only active items in same category).
	var neighborID int64
	var neighborOrder int
	var query string
	if moveUp {
		query = `SELECT id, sort_order FROM items WHERE category_id = $1 AND deleted_at IS NULL AND sort_order < $2 ORDER BY sort_order DESC LIMIT 1`
	} else {
		query = `SELECT id, sort_order FROM items WHERE category_id = $1 AND deleted_at IS NULL AND sort_order > $2 ORDER BY sort_order ASC LIMIT 1`
	}
	err = db.QueryRow(ctx, query, categoryID, currentOrder).Scan(&neighborID, &neighborOrder)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil // already at edge, no-op
		}
		return fmt.Errorf("reorder item find neighbor: %w", err)
	}

	// Swap sort_orders.
	_, err = db.Exec(ctx, `UPDATE items SET sort_order = $1 WHERE id = $2`, neighborOrder, id)
	if err != nil {
		return fmt.Errorf("reorder item update target: %w", err)
	}
	_, err = db.Exec(ctx, `UPDATE items SET sort_order = $1 WHERE id = $2`, currentOrder, neighborID)
	if err != nil {
		return fmt.Errorf("reorder item update neighbor: %w", err)
	}
	return nil
}

func reorderItemEdge(ctx context.Context, db DBTX, id int64, toTop bool) error {
	// Get current item's category_id and verify it exists.
	var categoryID int64
	err := db.QueryRow(ctx, `
		SELECT category_id FROM items
		WHERE id = $1 AND deleted_at IS NULL`, id).Scan(&categoryID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("reorder item get current: %w", err)
	}

	var newOrder int
	if toTop {
		err = db.QueryRow(ctx, `SELECT COALESCE(MIN(sort_order), 0) - 1 FROM items WHERE category_id = $1 AND deleted_at IS NULL`, categoryID).Scan(&newOrder)
	} else {
		err = db.QueryRow(ctx, `SELECT COALESCE(MAX(sort_order), 0) + 1 FROM items WHERE category_id = $1 AND deleted_at IS NULL`, categoryID).Scan(&newOrder)
	}
	if err != nil {
		return fmt.Errorf("reorder item calc edge: %w", err)
	}

	_, err = db.Exec(ctx, `UPDATE items SET sort_order = $1 WHERE id = $2`, newOrder, id)
	if err != nil {
		return fmt.Errorf("reorder item update: %w", err)
	}
	return nil
}
