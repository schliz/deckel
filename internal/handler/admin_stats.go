package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/schliz/deckel/internal/auth"
	"github.com/schliz/deckel/internal/middleware"
	"github.com/schliz/deckel/internal/model"
	"github.com/schliz/deckel/internal/store"
)

// StatsPageData is the view model for the admin statistics page.
type StatsPageData struct {
	User              *auth.RequestUser
	TotalRevenue      int64
	TotalDeposits     int64
	TransactionCount  int
	TopItemsByCount    []store.ItemStat
	TopItemsByRevenue  []store.ItemStat
	RevenueByCategory  []store.CategoryStat
	From              string
	To                string
	ActivePage        string
	LowBalanceWarning bool
	CSRFToken         string
}

// AdminStatsPage renders the admin statistics page for a selected time range.
func (h *Handler) AdminStatsPage(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	user := auth.UserFromContext(ctx)
	db := h.Store.DB()

	// Parse from/to query params; default to current month.
	now := time.Now()
	fromDefault := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	toDefault := fromDefault.AddDate(0, 1, 0)

	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	from := fromDefault
	to := toDefault

	if fromStr != "" {
		if parsed, err := time.ParseInLocation("2006-01-02", fromStr, now.Location()); err == nil {
			from = parsed
		}
	}
	if toStr != "" {
		if parsed, err := time.ParseInLocation("2006-01-02", toStr, now.Location()); err == nil {
			to = parsed
		}
	}

	totalRevenue, err := store.GetTotalRevenue(ctx, db, from, to)
	if err != nil {
		return fmt.Errorf("admin stats: get total revenue: %w", err)
	}

	totalDeposits, err := store.GetTotalDeposits(ctx, db, from, to)
	if err != nil {
		return fmt.Errorf("admin stats: get total deposits: %w", err)
	}

	txnCount, err := store.GetTransactionCount(ctx, db, from, to)
	if err != nil {
		return fmt.Errorf("admin stats: get transaction count: %w", err)
	}

	topByCount, err := store.GetTopItemsByCount(ctx, db, from, to, 10)
	if err != nil {
		return fmt.Errorf("admin stats: get top items by count: %w", err)
	}

	topByRevenue, err := store.GetTopItemsByRevenue(ctx, db, from, to, 10)
	if err != nil {
		return fmt.Errorf("admin stats: get top items by revenue: %w", err)
	}

	revenueByCategory, err := store.GetRevenueByCategory(ctx, db, from, to)
	if err != nil {
		return fmt.Errorf("admin stats: get revenue by category: %w", err)
	}

	// Fetch settings for low balance warning.
	var settings *model.Settings
	settings, err = store.GetSettings(ctx, db)
	if err != nil {
		return fmt.Errorf("admin stats: get settings: %w", err)
	}

	data := StatsPageData{
		User:              user,
		TotalRevenue:      totalRevenue,
		TotalDeposits:     totalDeposits,
		TransactionCount:  txnCount,
		TopItemsByCount:   topByCount,
		TopItemsByRevenue:  topByRevenue,
		RevenueByCategory:  revenueByCategory,
		From:              from.Format("2006-01-02"),
		To:                to.Format("2006-01-02"),
		ActivePage:        "admin-stats",
		LowBalanceWarning: isLowBalance(user, settings),
		CSRFToken:         middleware.CSRFTokenFromContext(ctx),
	}

	if isHTMX(r) {
		h.Renderer.Fragment(w, r, "admin-stats-panel", data)
		return nil
	}

	h.Renderer.Page(w, r, "admin_stats", data)
	return nil
}
