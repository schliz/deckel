package store

import (
	"context"
	"fmt"
	"time"
)

// ItemStat holds aggregated statistics for an item.
type ItemStat struct {
	Name    string
	Count   int
	Revenue int64
}

// GetTotalRevenue returns the sum of absolute values of negative non-cancelled transactions
// (i.e. purchases/spending) within the given time range.
func GetTotalRevenue(ctx context.Context, db DBTX, from, to time.Time) (int64, error) {
	var total int64
	err := db.QueryRow(ctx, `
		SELECT COALESCE(SUM(ABS(amount)), 0)
		FROM transactions
		WHERE amount < 0
		  AND cancelled_at IS NULL
		  AND type != 'cancellation'
		  AND created_at >= $1
		  AND created_at < $2`, from, to).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("get total revenue: %w", err)
	}
	return total, nil
}

// GetTopItemsByCount returns the top items by transaction count within the given time range.
func GetTopItemsByCount(ctx context.Context, db DBTX, from, to time.Time, limit int) ([]ItemStat, error) {
	rows, err := db.Query(ctx, `
		SELECT item_title, COUNT(*) AS cnt, COALESCE(SUM(ABS(amount)), 0) AS revenue
		FROM transactions
		WHERE item_title IS NOT NULL
		  AND cancelled_at IS NULL
		  AND type != 'cancellation'
		  AND created_at >= $1
		  AND created_at < $2
		GROUP BY item_title
		ORDER BY cnt DESC
		LIMIT $3`, from, to, limit)
	if err != nil {
		return nil, fmt.Errorf("get top items by count: %w", err)
	}
	defer rows.Close()

	var items []ItemStat
	for rows.Next() {
		var s ItemStat
		if err := rows.Scan(&s.Name, &s.Count, &s.Revenue); err != nil {
			return nil, fmt.Errorf("scan item stat: %w", err)
		}
		items = append(items, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate item stats: %w", err)
	}
	return items, nil
}

// GetTopItemsByRevenue returns the top items by total revenue within the given time range.
func GetTopItemsByRevenue(ctx context.Context, db DBTX, from, to time.Time, limit int) ([]ItemStat, error) {
	rows, err := db.Query(ctx, `
		SELECT item_title, COUNT(*) AS cnt, COALESCE(SUM(ABS(amount)), 0) AS revenue
		FROM transactions
		WHERE item_title IS NOT NULL
		  AND cancelled_at IS NULL
		  AND type != 'cancellation'
		  AND created_at >= $1
		  AND created_at < $2
		GROUP BY item_title
		ORDER BY revenue DESC
		LIMIT $3`, from, to, limit)
	if err != nil {
		return nil, fmt.Errorf("get top items by revenue: %w", err)
	}
	defer rows.Close()

	var items []ItemStat
	for rows.Next() {
		var s ItemStat
		if err := rows.Scan(&s.Name, &s.Count, &s.Revenue); err != nil {
			return nil, fmt.Errorf("scan item stat: %w", err)
		}
		items = append(items, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate item stats: %w", err)
	}
	return items, nil
}

// GetTransactionCount returns the count of non-cancelled transactions within the given time range.
func GetTransactionCount(ctx context.Context, db DBTX, from, to time.Time) (int, error) {
	var count int
	err := db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM transactions
		WHERE cancelled_at IS NULL
		  AND type != 'cancellation'
		  AND created_at >= $1
		  AND created_at < $2`, from, to).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("get transaction count: %w", err)
	}
	return count, nil
}

// CategoryStat holds aggregated statistics for a category.
type CategoryStat struct {
	Name    string
	Count   int
	Revenue int64
}

// GetRevenueByCategory returns purchase counts and revenue grouped by item category
// within the given time range. It maps transactions to categories via item_title,
// using DISTINCT ON to resolve any duplicate item names (preferring non-deleted items).
func GetRevenueByCategory(ctx context.Context, db DBTX, from, to time.Time) ([]CategoryStat, error) {
	rows, err := db.Query(ctx, `
		SELECT c.name, COUNT(*) AS cnt, COALESCE(SUM(ABS(t.amount)), 0) AS revenue
		FROM transactions t
		JOIN (
			SELECT DISTINCT ON (name) name, category_id
			FROM items
			ORDER BY name, deleted_at NULLS FIRST
		) i ON i.name = t.item_title
		JOIN categories c ON c.id = i.category_id
		WHERE t.item_title IS NOT NULL
		  AND t.cancelled_at IS NULL
		  AND t.type != 'cancellation'
		  AND t.created_at >= $1
		  AND t.created_at < $2
		GROUP BY c.id, c.name
		ORDER BY revenue DESC`, from, to)
	if err != nil {
		return nil, fmt.Errorf("get revenue by category: %w", err)
	}
	defer rows.Close()

	var stats []CategoryStat
	for rows.Next() {
		var s CategoryStat
		if err := rows.Scan(&s.Name, &s.Count, &s.Revenue); err != nil {
			return nil, fmt.Errorf("scan category stat: %w", err)
		}
		stats = append(stats, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate category stats: %w", err)
	}
	return stats, nil
}

// GetTotalDeposits returns the sum of deposit transactions within the given time range.
func GetTotalDeposits(ctx context.Context, db DBTX, from, to time.Time) (int64, error) {
	var total int64
	err := db.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount), 0)
		FROM transactions
		WHERE type = 'deposit'
		  AND cancelled_at IS NULL
		  AND created_at >= $1
		  AND created_at < $2`, from, to).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("get total deposits: %w", err)
	}
	return total, nil
}
