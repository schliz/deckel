package member

// Member custom-transaction mutations: the create-custom-transaction modal,
// the cancel-transaction modal, plus the handlers that book and void
// transactions. Reuses TransactionHistoryData from history.go to refresh the
// transaction list after a cancellation.

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/schliz/deckel/internal/auth"
	"github.com/schliz/deckel/internal/handler"
	"github.com/schliz/deckel/internal/middleware"
	"github.com/schliz/deckel/internal/model"
	"github.com/schliz/deckel/internal/store"
)

// CustomTransactionModalData is the view model for the custom transaction modal.
type CustomTransactionModalData struct {
	Settings     *model.Settings
	CSRFToken    string
	ErrorMessage string
}

// CustomTransactionModal renders the custom transaction modal.
func (h *Handler) CustomTransactionModal(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	db := h.Store.DB()

	settings, err := store.GetSettings(ctx, db)
	if err != nil {
		return fmt.Errorf("custom transaction modal: get settings: %w", err)
	}

	data := CustomTransactionModalData{
		Settings:  settings,
		CSRFToken: middleware.CSRFTokenFromContext(ctx),
	}

	h.Renderer.Fragment(w, r, "custom-transaction-modal", data)
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
		return &handler.NotFoundError{Message: "Transaktion nicht gefunden"}
	}

	// Fetch transaction.
	txn, err := store.GetTransaction(ctx, db, id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return &handler.NotFoundError{Message: "Transaktion nicht gefunden"}
		}
		return fmt.Errorf("cancel modal: get transaction: %w", err)
	}

	// Verify transaction belongs to current user.
	if txn.UserID != user.ID {
		return &handler.ForbiddenError{Message: "Zugriff verweigert"}
	}

	// Only self-created transactions can be cancelled by the user.
	if txn.CreatedByUserID == nil || *txn.CreatedByUserID != user.ID {
		return &handler.ForbiddenError{Message: "Diese Transaktion kann nur vom Ersteller storniert werden."}
	}

	// Verify not already cancelled.
	if txn.CancelledAt != nil {
		return &handler.ValidationError{Message: "Transaktion wurde bereits storniert"}
	}

	// Stornobuchungen cannot be voided.
	if txn.Type == "cancellation" {
		return &handler.ValidationError{Message: "Stornobuchungen können nicht storniert werden"}
	}

	// Fetch settings for cancellation window.
	settings, err := store.GetSettings(ctx, db)
	if err != nil {
		return fmt.Errorf("cancel modal: get settings: %w", err)
	}

	// Verify within cancellation window.
	if time.Since(txn.CreatedAt) > time.Duration(settings.CancellationMinutes)*time.Minute {
		return &handler.ValidationError{Message: "Stornierungsfenster abgelaufen"}
	}

	data := CancelModalData{
		Transaction: txn,
		CSRFToken:   middleware.CSRFTokenFromContext(ctx),
	}

	h.Renderer.Fragment(w, r, "cancel-modal", data)
	return nil
}

// CreateCustomTransaction processes a custom transaction submission.
func (h *Handler) CreateCustomTransaction(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	user := auth.UserFromContext(ctx)
	db := h.Store.DB()

	// Fetch settings early — needed for both validation and error re-rendering.
	settings, err := store.GetSettings(ctx, db)
	if err != nil {
		return fmt.Errorf("create custom transaction: get settings: %w", err)
	}

	// Re-render the modal with an inline error message.
	renderError := func(msg string) error {
		data := CustomTransactionModalData{
			Settings:     settings,
			CSRFToken:    middleware.CSRFTokenFromContext(ctx),
			ErrorMessage: msg,
		}
		h.Renderer.Fragment(w, r, "custom-transaction-modal", data)
		return nil
	}

	// Parse form data.
	description := strings.TrimSpace(r.FormValue("description"))

	if description == "" {
		return renderError("Beschreibung ist erforderlich")
	}
	if err := handler.ValidateTextLen(description, 500, "Beschreibung"); err != nil {
		var ve *handler.ValidationError
		if errors.As(err, &ve) {
			return renderError(ve.Message)
		}
		return err
	}

	// Parse Euro amount string to cents (accepts both comma and period).
	amountFloat, err := strconv.ParseFloat(handler.NormalizeDecimal(r.FormValue("amount")), 64)
	if err != nil || amountFloat <= 0 {
		return renderError("Ungültiger Betrag")
	}
	amountCents := int64(math.Round(amountFloat * 100))

	// Validate amount within allowed range.
	if amountCents < settings.CustomTxMin {
		return renderError(fmt.Sprintf("Mindestbetrag: %s", formatEuroCents(settings.CustomTxMin)))
	}
	if amountCents > settings.CustomTxMax {
		return renderError(fmt.Sprintf("Maximalbetrag: %s", formatEuroCents(settings.CustomTxMax)))
	}

	// Create transaction (amount is negative = debit).
	createdBy := user.ID
	txn := &model.Transaction{
		UserID:          user.ID,
		Amount:          -amountCents,
		Description:     &description,
		Type:            "custom",
		CreatedByUserID: &createdBy,
	}

	err = h.Store.WithTx(ctx, func(tx pgx.Tx) error {
		balance, err := store.GetUserBalanceForUpdate(ctx, tx, user.ID)
		if err != nil {
			return fmt.Errorf("get balance for update: %w", err)
		}

		if settings.HardLimitEnabled && !user.SpendingLimitDisabled {
			if balance <= -settings.HardSpendingLimit {
				return &handler.ValidationError{Message: "Buchung nicht möglich: Ausgabenlimit erreicht. Bitte erst einzahlen."}
			}
		}

		_, err = store.CreateTransaction(ctx, tx, txn)
		return err
	})
	if err != nil {
		var valErr *handler.ValidationError
		if errors.As(err, &valErr) {
			return valErr
		}
		return fmt.Errorf("create custom transaction: %w", err)
	}

	// Render OOB header-stats update.
	newBalance, _ := store.GetUserBalance(ctx, db, user.ID)

	// Show warning overlay if the user's balance is now below the warning limit.
	warning := newBalance < settings.WarningLimit

	// Build response: success overlay + OOB header-stats.
	h.Renderer.Fragment(w, r, "order-success", map[string]any{
		"Title":       "Buchung gespeichert!",
		"Subtitle":    description,
		"TotalAmount": amountCents,
		"Warning":     warning,
	})
	totalBalance, _ := store.GetAllBalancesSum(ctx, db)
	negativeSum, _ := store.GetNegativeBalancesSum(ctx, db)
	rank, total, _ := store.GetUserRank(ctx, db, user.ID)

	h.Renderer.AppendOOB(w, "header-stats", map[string]any{
		"UserBalance":         newBalance,
		"TotalBalance":        totalBalance,
		"NegativeBalancesSum": negativeSum,
		"UserRank":            rank,
		"TotalUsers":          total,
		"Settings":            settings,
		"User":                user,
		"OOB":                 true,
	})

	return nil
}

// formatEuroCents formats cents as Euro string (e.g., 150 -> "1,50 EUR").
func formatEuroCents(cents int64) string {
	euros := cents / 100
	remainder := cents % 100
	if remainder < 0 {
		remainder = -remainder
	}
	return fmt.Sprintf("%d,%02d EUR", euros, remainder)
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
		return &handler.NotFoundError{Message: "Transaktion nicht gefunden"}
	}

	// Fetch transaction.
	txn, err := store.GetTransaction(ctx, db, id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return &handler.NotFoundError{Message: "Transaktion nicht gefunden"}
		}
		return fmt.Errorf("cancel transaction: get transaction: %w", err)
	}

	// Verify transaction belongs to current user.
	if txn.UserID != user.ID {
		return &handler.ForbiddenError{Message: "Zugriff verweigert"}
	}

	// Only self-created transactions can be cancelled by the user.
	if txn.CreatedByUserID == nil || *txn.CreatedByUserID != user.ID {
		return &handler.ForbiddenError{Message: "Diese Transaktion kann nur vom Ersteller storniert werden."}
	}

	// Verify not already cancelled.
	if txn.CancelledAt != nil {
		return &handler.ValidationError{Message: "Transaktion wurde bereits storniert"}
	}

	// Stornobuchungen cannot be voided.
	if txn.Type == "cancellation" {
		return &handler.ValidationError{Message: "Stornobuchungen können nicht storniert werden"}
	}

	// Fetch settings for cancellation window.
	settings, err := store.GetSettings(ctx, db)
	if err != nil {
		return fmt.Errorf("cancel transaction: get settings: %w", err)
	}

	// Verify within cancellation window.
	if time.Since(txn.CreatedAt) > time.Duration(settings.CancellationMinutes)*time.Minute {
		return &handler.ValidationError{Message: "Stornierungsfenster abgelaufen"}
	}

	// Execute cancellation within a DB transaction.
	err = h.Store.WithTx(ctx, func(tx pgx.Tx) error {
		return store.CancelTransaction(ctx, tx, id, user.ID)
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
	negativeSum, _ := store.GetNegativeBalancesSum(ctx, db)
	rank, total, _ := store.GetUserRank(ctx, db, user.ID)

	h.Renderer.AppendOOB(w, "header-stats", map[string]any{
		"UserBalance":         newBalance,
		"TotalBalance":        totalBalance,
		"NegativeBalancesSum": negativeSum,
		"UserRank":            rank,
		"TotalUsers":          total,
		"Settings":            settings,
		"User":                user,
		"OOB":                 true,
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
