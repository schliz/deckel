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

// TransactionHistoryData is the view model for the transaction history page.
type TransactionHistoryData struct {
	User                *auth.RequestUser
	Transactions        []model.Transaction
	Settings            *model.Settings
	CSRFToken           string
	ActivePage          string
	CancellationMinutes int
	Page                int
	TotalPages          int
}

// TransactionHistory renders the paginated transaction history for the current user.
func (h *Handler) TransactionHistory(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	user := auth.UserFromContext(ctx)
	db := h.Store.DB()

	// Fetch settings for pagination_size and cancellation_minutes.
	settings, err := store.GetSettings(ctx, db)
	if err != nil {
		return fmt.Errorf("transaction history: get settings: %w", err)
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

	// Fetch paginated transactions for the current user.
	txns, total, err := store.ListTransactionsByUser(ctx, db, user.ID, limit, offset)
	if err != nil {
		return fmt.Errorf("transaction history: list transactions: %w", err)
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

	data := TransactionHistoryData{
		User:                user,
		Transactions:        txns,
		Settings:            settings,
		CSRFToken:           middleware.CSRFTokenFromContext(ctx),
		ActivePage:          "transactions",
		CancellationMinutes: settings.CancellationMinutes,
		Page:                page,
		TotalPages:          totalPages,
	}

	h.Renderer.Page(w, r, "history", data)
	return nil
}
