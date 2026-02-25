package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
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

// CancelModalData is the view model for the cancel confirmation modal.
type CancelModalData struct {
	Transaction *model.Transaction
	CSRFToken   string
}

// CancelModal renders the cancel confirmation modal for a given transaction.
func (h *Handler) CancelModal(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	user := auth.UserFromContext(ctx)
	db := h.Store.DB()

	// Extract transaction ID from URL path.
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return &NotFoundError{Message: "Transaktion nicht gefunden"}
	}

	// Fetch transaction.
	txn, err := store.GetTransaction(ctx, db, id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return &NotFoundError{Message: "Transaktion nicht gefunden"}
		}
		return fmt.Errorf("cancel modal: get transaction: %w", err)
	}

	// Verify transaction belongs to current user.
	if txn.UserID != user.ID {
		return &ForbiddenError{Message: "Zugriff verweigert"}
	}

	// Verify not already cancelled.
	if txn.CancelledAt != nil {
		return &ValidationError{Message: "Transaktion wurde bereits storniert"}
	}

	// Fetch settings for cancellation window.
	settings, err := store.GetSettings(ctx, db)
	if err != nil {
		return fmt.Errorf("cancel modal: get settings: %w", err)
	}

	// Verify within cancellation window.
	if time.Since(txn.CreatedAt) > time.Duration(settings.CancellationMinutes)*time.Minute {
		return &ValidationError{Message: "Stornierungsfenster abgelaufen"}
	}

	data := CancelModalData{
		Transaction: txn,
		CSRFToken:   middleware.CSRFTokenFromContext(ctx),
	}

	h.Renderer.Fragment(w, r, "cancel-modal", data)
	return nil
}

// CancelTransaction processes the cancellation of a transaction.
func (h *Handler) CancelTransaction(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	user := auth.UserFromContext(ctx)
	db := h.Store.DB()

	// Extract transaction ID from URL path.
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return &NotFoundError{Message: "Transaktion nicht gefunden"}
	}

	// Fetch transaction.
	txn, err := store.GetTransaction(ctx, db, id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return &NotFoundError{Message: "Transaktion nicht gefunden"}
		}
		return fmt.Errorf("cancel transaction: get transaction: %w", err)
	}

	// Verify transaction belongs to current user.
	if txn.UserID != user.ID {
		return &ForbiddenError{Message: "Zugriff verweigert"}
	}

	// Verify not already cancelled.
	if txn.CancelledAt != nil {
		return &ValidationError{Message: "Transaktion wurde bereits storniert"}
	}

	// Fetch settings for cancellation window.
	settings, err := store.GetSettings(ctx, db)
	if err != nil {
		return fmt.Errorf("cancel transaction: get settings: %w", err)
	}

	// Verify within cancellation window.
	if time.Since(txn.CreatedAt) > time.Duration(settings.CancellationMinutes)*time.Minute {
		return &ValidationError{Message: "Stornierungsfenster abgelaufen"}
	}

	// Execute cancellation within a DB transaction.
	err = h.Store.WithTx(ctx, func(tx pgx.Tx) error {
		return store.CancelTransaction(ctx, tx, id)
	})
	if err != nil {
		return fmt.Errorf("cancel transaction: %w", err)
	}

	// Build response: toast + OOB header-stats + swapped transaction list.

	// Render success toast.
	h.Renderer.Fragment(w, r, "toast", map[string]string{
		"Type":    "success",
		"Message": "Transaktion storniert!",
	})

	// Render OOB header-stats update.
	newBalance, _ := store.GetUserBalance(ctx, db, user.ID)
	totalBalance, _ := store.GetAllBalancesSum(ctx, db)
	rank, total, _ := store.GetUserRank(ctx, db, user.ID)

	h.Renderer.AppendOOB(w, "header-stats", map[string]any{
		"UserBalance":  newBalance,
		"TotalBalance": totalBalance,
		"UserRank":     rank,
		"TotalUsers":   total,
		"Settings":     settings,
	})

	// Re-render the transaction list for the current page.
	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	limit := settings.PaginationSize
	offset := (page - 1) * limit

	txns, totalTxns, err := store.ListTransactionsByUser(ctx, db, user.ID, limit, offset)
	if err != nil {
		return fmt.Errorf("cancel transaction: list transactions: %w", err)
	}
	totalPages := 0
	if totalTxns > 0 {
		totalPages = (totalTxns + limit - 1) / limit
	}

	h.Renderer.AppendOOB(w, "transaction-list", TransactionHistoryData{
		User:                user,
		Transactions:        txns,
		Settings:            settings,
		CSRFToken:           middleware.CSRFTokenFromContext(ctx),
		ActivePage:          "transactions",
		CancellationMinutes: settings.CancellationMinutes,
		Page:                page,
		TotalPages:          totalPages,
	})

	// Close modal.
	w.Write([]byte(`<div id="modal" hx-swap-oob="innerHTML" style="display:none"></div>`))

	return nil
}
