package member

// Member transaction history: read-only paginated view of the current user's
// transactions. View-model struct lives here because the cancel handler in
// transactions.go also reuses it for OOB list refreshes.

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/schliz/deckel/internal/auth"
	"github.com/schliz/deckel/internal/handler"
	"github.com/schliz/deckel/internal/middleware"
	"github.com/schliz/deckel/internal/model"
	"github.com/schliz/deckel/internal/store"
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
	LowBalanceWarning   bool
	IsBlocked           bool
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

	// Determine if user is blocked (at or below hard spending limit).
	isBlocked := false
	if settings.HardLimitEnabled && user != nil && !user.SpendingLimitDisabled {
		isBlocked = user.Balance <= -settings.HardSpendingLimit
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
		LowBalanceWarning:   handler.IsLowBalance(user, settings),
		IsBlocked:           isBlocked,
	}

	if handler.IsHTMX(r) {
		h.Renderer.Fragment(w, r, "transaction-list", data)
		return nil
	}

	h.Renderer.Page(w, r, "member/history", data)
	return nil
}
