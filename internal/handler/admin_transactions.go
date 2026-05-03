package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/schliz/deckel/internal/auth"
	"github.com/schliz/deckel/internal/middleware"
	"github.com/schliz/deckel/internal/model"
	"github.com/schliz/deckel/internal/store"
)

// AdminTransactionsPageData is the view model for the admin transaction list page.
type AdminTransactionsPageData struct {
	User              *auth.RequestUser
	Transactions      []model.TransactionWithUser
	Settings          *model.Settings
	CSRFToken         string
	ActivePage        string
	Page              int
	TotalPages        int
	LowBalanceWarning bool
}

// AdminTransactionList renders the paginated admin transaction list (newest first).
func (h *Base) AdminTransactionList(w http.ResponseWriter, r *http.Request) error {
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
		User:              user,
		Transactions:      transactions,
		Settings:          settings,
		CSRFToken:         middleware.CSRFTokenFromContext(ctx),
		ActivePage:        "admin-transactions",
		Page:              page,
		TotalPages:        totalPages,
		LowBalanceWarning: IsLowBalance(user, settings),
	}

	if IsHTMX(r) {
		h.Renderer.Fragment(w, r, "admin-transaction-list", data)
		return nil
	}

	h.Renderer.Page(w, r, "admin_transactions", data)
	return nil
}

// AdminCancelModalData is the view model for the admin cancel confirmation modal.
type AdminCancelModalData struct {
	Transaction *model.Transaction
	UserName    string
	CSRFToken   string
}

// AdminCancelModal renders the cancel confirmation modal for an admin cancelling any transaction.
func (h *Base) AdminCancelModal(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	db := h.Store.DB()

	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return &NotFoundError{Message: "Transaktion nicht gefunden"}
	}

	txn, err := store.GetTransaction(ctx, db, id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return &NotFoundError{Message: "Transaktion nicht gefunden"}
		}
		return fmt.Errorf("admin cancel modal: get transaction: %w", err)
	}

	if txn.CancelledAt != nil {
		return &ValidationError{Message: "Transaktion wurde bereits storniert"}
	}

	if txn.Type == "cancellation" {
		return &ValidationError{Message: "Stornobuchungen können nicht storniert werden"}
	}

	settings, err := store.GetSettings(ctx, db)
	if err != nil {
		return fmt.Errorf("admin cancel modal: get settings: %w", err)
	}

	if time.Since(txn.CreatedAt) > time.Duration(settings.CancellationMinutes)*time.Minute {
		return &ValidationError{Message: "Stornierungsfenster abgelaufen"}
	}

	// Get user name for display.
	var userName string
	err = db.QueryRow(ctx, `SELECT full_name FROM users WHERE id = $1`, txn.UserID).Scan(&userName)
	if err != nil {
		return fmt.Errorf("admin cancel modal: get user name: %w", err)
	}

	data := AdminCancelModalData{
		Transaction: txn,
		UserName:    userName,
		CSRFToken:   middleware.CSRFTokenFromContext(ctx),
	}

	h.Renderer.Fragment(w, r, "admin-cancel-modal", data)
	return nil
}

// AdminCancelTransaction processes the cancellation of any transaction by an admin.
func (h *Base) AdminCancelTransaction(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	user := auth.UserFromContext(ctx)
	db := h.Store.DB()

	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return &NotFoundError{Message: "Transaktion nicht gefunden"}
	}

	txn, err := store.GetTransaction(ctx, db, id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return &NotFoundError{Message: "Transaktion nicht gefunden"}
		}
		return fmt.Errorf("admin cancel transaction: get transaction: %w", err)
	}

	if txn.CancelledAt != nil {
		return &ValidationError{Message: "Transaktion wurde bereits storniert"}
	}

	if txn.Type == "cancellation" {
		return &ValidationError{Message: "Stornobuchungen können nicht storniert werden"}
	}

	settings, err := store.GetSettings(ctx, db)
	if err != nil {
		return fmt.Errorf("admin cancel transaction: get settings: %w", err)
	}

	if time.Since(txn.CreatedAt) > time.Duration(settings.CancellationMinutes)*time.Minute {
		return &ValidationError{Message: "Stornierungsfenster abgelaufen"}
	}

	err = h.Store.WithTx(ctx, func(tx pgx.Tx) error {
		return store.CancelTransaction(ctx, tx, id, user.ID)
	})
	if err != nil {
		return fmt.Errorf("admin cancel transaction: %w", err)
	}

	// Success toast.
	h.Renderer.Fragment(w, r, "toast", map[string]string{
		"Type":    "success",
		"Message": "Transaktion storniert!",
	})

	// Render OOB header-stats update (total balance changes on cancel).
	reqUser := auth.UserFromContext(ctx)
	newBalance, _ := store.GetUserBalance(ctx, db, reqUser.ID)
	totalBalance, _ := store.GetAllBalancesSum(ctx, db)
	negativeSum, _ := store.GetNegativeBalancesSum(ctx, db)
	rank, total, _ := store.GetUserRank(ctx, db, reqUser.ID)

	h.Renderer.AppendOOB(w, "header-stats", map[string]any{
		"UserBalance":         newBalance,
		"TotalBalance":        totalBalance,
		"NegativeBalancesSum": negativeSum,
		"UserRank":            rank,
		"TotalUsers":          total,
		"Settings":            settings,
		"User":                reqUser,
		"OOB":                 true,
	})

	// Re-render the admin transaction list for the current page.
	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	limit := settings.PaginationSize
	offset := (page - 1) * limit

	transactions, totalTxns, err := store.ListAllTransactions(ctx, db, limit, offset)
	if err != nil {
		return fmt.Errorf("admin cancel transaction: list transactions: %w", err)
	}
	totalPages := 0
	if totalTxns > 0 {
		totalPages = (totalTxns + limit - 1) / limit
	}

	h.Renderer.AppendOOB(w, "admin-transaction-list", AdminTransactionsPageData{
		User:         user,
		Transactions: transactions,
		Settings:     settings,
		CSRFToken:    middleware.CSRFTokenFromContext(ctx),
		ActivePage:   "admin-transactions",
		Page:         page,
		TotalPages:   totalPages,
	})

	// Close modal.
	w.Write([]byte(`<div id="modal" hx-swap-oob="innerHTML" style="display:none"></div>`))

	return nil
}
