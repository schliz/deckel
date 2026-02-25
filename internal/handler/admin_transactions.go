package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/k4-bar/deckel/internal/auth"
	"github.com/k4-bar/deckel/internal/middleware"
	"github.com/k4-bar/deckel/internal/model"
	"github.com/k4-bar/deckel/internal/store"
)

// AdminTransactionsPageData is the view model for the admin transaction list page.
type AdminTransactionsPageData struct {
	User         *auth.RequestUser
	Transactions []model.TransactionWithUser
	Settings     *model.Settings
	CSRFToken    string
	ActivePage   string
	Page         int
	TotalPages   int
}

// AdminTransactionList renders the paginated admin transaction list (newest first).
func (h *Handler) AdminTransactionList(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	user := auth.UserFromContext(ctx)
	db := h.Store.DB()

	// Fetch settings for pagination size.
	settings, err := store.GetSettings(ctx, db)
	if err != nil {
		return fmt.Errorf("admin transaction list: get settings: %w", err)
	}

	// Read page query param (default 1).
	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	limit := settings.PaginationSize
	offset := (page - 1) * limit

	// Fetch paginated transactions with user info.
	transactions, total, err := store.ListAllTransactions(ctx, db, limit, offset)
	if err != nil {
		return fmt.Errorf("admin transaction list: list transactions: %w", err)
	}

	// Compute total pages.
	totalPages := 0
	if total > 0 {
		totalPages = (total + limit - 1) / limit
	}

	// Clamp page to valid range.
	if page > totalPages && totalPages > 0 {
		page = totalPages
	}

	data := AdminTransactionsPageData{
		User:         user,
		Transactions: transactions,
		Settings:     settings,
		CSRFToken:    middleware.CSRFTokenFromContext(ctx),
		ActivePage:   "admin-transactions",
		Page:         page,
		TotalPages:   totalPages,
	}

	h.Renderer.Page(w, r, "admin_transactions", data)
	return nil
}
