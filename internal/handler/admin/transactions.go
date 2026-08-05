package admin

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/schliz/deckel/internal/auth"
	"github.com/schliz/deckel/internal/handler"
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
		User:              user,
		Transactions:      transactions,
		Settings:          settings,
		CSRFToken:         middleware.CSRFTokenFromContext(ctx),
		ActivePage:        "admin-transactions",
		Page:              page,
		TotalPages:        totalPages,
		LowBalanceWarning: handler.IsLowBalance(user, settings),
	}

	if handler.IsHTMX(r) {
		h.Renderer.Fragment(w, r, "admin-transaction-list", data)
		return nil
	}

	h.Renderer.Page(w, r, "admin/transactions", data)
	return nil
}

// AdminUserTransactionsPageData is the view model for the admin single-user
// transaction list page, which also surfaces the synthetic "Anfangsbestand"
// (initial balance) carried over from anonymized transactions.
type AdminUserTransactionsPageData struct {
	User              *auth.RequestUser
	TargetUser        *model.UserWithBalance
	Transactions      []model.Transaction
	Settings          *model.Settings
	CSRFToken         string
	ActivePage        string
	Page              int
	TotalPages        int
	LowBalanceWarning bool
	ShowCarryover     bool
}

// AdminUserTransactionList renders the paginated transaction list for a single user.
func (h *Handler) AdminUserTransactionList(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	user := auth.UserFromContext(ctx)
	db := h.Store.DB()

	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return &handler.NotFoundError{Message: "Benutzer nicht gefunden"}
	}

	targetUser, err := store.GetUserWithBalance(ctx, db, id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return &handler.NotFoundError{Message: "Benutzer nicht gefunden"}
		}
		return fmt.Errorf("admin user transaction list: get user with balance: %w", err)
	}

	// Fetch settings for pagination size.
	settings, err := store.GetSettings(ctx, db)
	if err != nil {
		return fmt.Errorf("admin user transaction list: get settings: %w", err)
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

	// Fetch paginated transactions for the target user.
	transactions, total, err := store.ListTransactionsByUser(ctx, db, id, limit, offset)
	if err != nil {
		return fmt.Errorf("admin user transaction list: list transactions: %w", err)
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

	// The synthetic "Anfangsbestand" row only makes sense on the oldest page
	// (transactions are newest-first), and only if there is a carryover to show.
	showCarryover := targetUser.BalanceCarryover != 0 && (totalPages == 0 || page >= totalPages)

	data := AdminUserTransactionsPageData{
		User:              user,
		TargetUser:        targetUser,
		Transactions:      transactions,
		Settings:          settings,
		CSRFToken:         middleware.CSRFTokenFromContext(ctx),
		ActivePage:        "admin-users",
		Page:              page,
		TotalPages:        totalPages,
		LowBalanceWarning: handler.IsLowBalance(user, settings),
		ShowCarryover:     showCarryover,
	}

	if handler.IsHTMX(r) {
		h.Renderer.Fragment(w, r, "admin-user-transaction-list", data)
		return nil
	}

	h.Renderer.Page(w, r, "admin/user_transactions", data)
	return nil
}

// AdminCancelModalData is the view model for the admin cancel confirmation modal.
type AdminCancelModalData struct {
	Transaction *model.Transaction
	UserName    string
	CSRFToken   string
	ListTarget  string
	PostURL     string
}

// AdminCancelModal renders the cancel confirmation modal for an admin cancelling any transaction.
func (h *Handler) AdminCancelModal(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	db := h.Store.DB()

	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return &handler.NotFoundError{Message: "Transaktion nicht gefunden"}
	}

	txn, err := store.GetTransaction(ctx, db, id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return &handler.NotFoundError{Message: "Transaktion nicht gefunden"}
		}
		return fmt.Errorf("admin cancel modal: get transaction: %w", err)
	}

	if txn.CancelledAt != nil {
		return &handler.ValidationError{Message: "Transaktion wurde bereits storniert"}
	}

	if txn.Type == "cancellation" {
		return &handler.ValidationError{Message: "Stornobuchungen können nicht storniert werden"}
	}

	settings, err := store.GetSettings(ctx, db)
	if err != nil {
		return fmt.Errorf("admin cancel modal: get settings: %w", err)
	}

	if time.Since(txn.CreatedAt) > time.Duration(settings.CancellationMinutes)*time.Minute {
		return &handler.ValidationError{Message: "Stornierungsfenster abgelaufen"}
	}

	if txn.UserID == nil {
		return &handler.ValidationError{Message: "Transaktion ist anonymisiert"}
	}

	// Get user name for display.
	var userName string
	err = db.QueryRow(ctx, `SELECT full_name FROM users WHERE id = $1`, *txn.UserID).Scan(&userName)
	if err != nil {
		return fmt.Errorf("admin cancel modal: get user name: %w", err)
	}

	// Read page query param (default 1).
	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	// The cancel confirmation is used both from the global admin transaction
	// list and from a single user's transaction list; the origin query param
	// tells us which list to target and refresh on submit.
	listTarget := "#admin-transaction-list"
	postURL := fmt.Sprintf("/admin/transactions/%d/cancel", id)
	if r.URL.Query().Get("origin") == "user" {
		listTarget = "#admin-user-transaction-list"
		postURL = fmt.Sprintf("/admin/transactions/%d/cancel?origin=user&page=%d", id, page)
	}

	data := AdminCancelModalData{
		Transaction: txn,
		UserName:    userName,
		CSRFToken:   middleware.CSRFTokenFromContext(ctx),
		ListTarget:  listTarget,
		PostURL:     postURL,
	}

	h.Renderer.Fragment(w, r, "admin-cancel-modal", data)
	return nil
}

// AdminCancelTransaction processes the cancellation of any transaction by an admin.
func (h *Handler) AdminCancelTransaction(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	user := auth.UserFromContext(ctx)
	db := h.Store.DB()

	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return &handler.NotFoundError{Message: "Transaktion nicht gefunden"}
	}

	txn, err := store.GetTransaction(ctx, db, id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return &handler.NotFoundError{Message: "Transaktion nicht gefunden"}
		}
		return fmt.Errorf("admin cancel transaction: get transaction: %w", err)
	}

	if txn.CancelledAt != nil {
		return &handler.ValidationError{Message: "Transaktion wurde bereits storniert"}
	}

	if txn.Type == "cancellation" {
		return &handler.ValidationError{Message: "Stornobuchungen können nicht storniert werden"}
	}

	settings, err := store.GetSettings(ctx, db)
	if err != nil {
		return fmt.Errorf("admin cancel transaction: get settings: %w", err)
	}

	if time.Since(txn.CreatedAt) > time.Duration(settings.CancellationMinutes)*time.Minute {
		return &handler.ValidationError{Message: "Stornierungsfenster abgelaufen"}
	}

	err = h.Store.WithTx(ctx, func(tx pgx.Tx) error {
		return store.CancelTransaction(ctx, tx, id, user.ID)
	})
	if err != nil {
		if errors.Is(err, store.ErrAnonymized) {
			return &handler.ValidationError{Message: "Transaktion ist anonymisiert"}
		}
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

	// Re-render the transaction list for the current page. The origin query
	// param determines whether this was triggered from the global admin
	// transaction list or from a single user's transaction list.
	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	limit := settings.PaginationSize
	offset := (page - 1) * limit

	if r.URL.Query().Get("origin") == "user" {
		userID := *txn.UserID

		targetUser, err := store.GetUserWithBalance(ctx, db, userID)
		if err != nil {
			return fmt.Errorf("admin cancel transaction: get user with balance: %w", err)
		}

		transactions, totalTxns, err := store.ListTransactionsByUser(ctx, db, userID, limit, offset)
		if err != nil {
			return fmt.Errorf("admin cancel transaction: list transactions by user: %w", err)
		}
		totalPages := 0
		if totalTxns > 0 {
			totalPages = (totalTxns + limit - 1) / limit
		}

		showCarryover := targetUser.BalanceCarryover != 0 && (totalPages == 0 || page >= totalPages)

		h.Renderer.AppendOOB(w, "admin-user-transaction-list", AdminUserTransactionsPageData{
			User:          user,
			TargetUser:    targetUser,
			Transactions:  transactions,
			Settings:      settings,
			CSRFToken:     middleware.CSRFTokenFromContext(ctx),
			ActivePage:    "admin-users",
			Page:          page,
			TotalPages:    totalPages,
			ShowCarryover: showCarryover,
		})
	} else {
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
	}

	// Close modal.
	w.Write([]byte(`<div id="modal" hx-swap-oob="innerHTML" style="display:none"></div>`))

	return nil
}
