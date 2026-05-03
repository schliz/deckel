package kiosk

// Cancellation flow for kiosk-created transactions: confirmation modal plus
// the cancel mutation that re-renders the history list.

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

// KioskCancelModal renders the cancel confirmation modal for a kiosk transaction.
func (h *Handler) KioskCancelModal(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	user := auth.UserFromContext(ctx)
	db := h.Store.DB()

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		return &handler.NotFoundError{Message: "Transaktion nicht gefunden"}
	}

	txn, err := store.GetTransaction(ctx, db, id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return &handler.NotFoundError{Message: "Transaktion nicht gefunden"}
		}
		return fmt.Errorf("kiosk cancel modal: get transaction: %w", err)
	}

	// Verify this transaction was created by the kiosk user.
	if txn.CreatedByUserID == nil || *txn.CreatedByUserID != user.ID {
		return &handler.ForbiddenError{Message: "Zugriff verweigert"}
	}

	if txn.CancelledAt != nil {
		return &handler.ValidationError{Message: "Transaktion wurde bereits storniert"}
	}

	if txn.Type == "cancellation" {
		return &handler.ValidationError{Message: "Stornobuchungen können nicht storniert werden"}
	}

	settings, err := store.GetSettings(ctx, db)
	if err != nil {
		return fmt.Errorf("kiosk cancel modal: get settings: %w", err)
	}

	if time.Since(txn.CreatedAt) > time.Duration(settings.CancellationMinutes)*time.Minute {
		return &handler.ValidationError{Message: "Stornierungsfenster abgelaufen"}
	}

	// We need the user name for the modal display - fetch it.
	targetUser, err := store.GetUserWithBalance(ctx, db, txn.UserID)
	if err != nil {
		return fmt.Errorf("kiosk cancel modal: get target user: %w", err)
	}

	data := map[string]any{
		"Transaction": struct {
			*model.Transaction
			UserName string
		}{txn, targetUser.FullName},
		"CSRFToken": middleware.CSRFTokenFromContext(ctx),
	}

	h.Renderer.Fragment(w, r, "kiosk-cancel-modal", data)
	return nil
}

// KioskCancelTransaction processes the cancellation of a kiosk-created transaction.
func (h *Handler) KioskCancelTransaction(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	user := auth.UserFromContext(ctx)
	db := h.Store.DB()

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		return &handler.NotFoundError{Message: "Transaktion nicht gefunden"}
	}

	txn, err := store.GetTransaction(ctx, db, id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return &handler.NotFoundError{Message: "Transaktion nicht gefunden"}
		}
		return fmt.Errorf("kiosk cancel: get transaction: %w", err)
	}

	if txn.CreatedByUserID == nil || *txn.CreatedByUserID != user.ID {
		return &handler.ForbiddenError{Message: "Zugriff verweigert"}
	}

	if txn.CancelledAt != nil {
		return &handler.ValidationError{Message: "Transaktion wurde bereits storniert"}
	}

	if txn.Type == "cancellation" {
		return &handler.ValidationError{Message: "Stornobuchungen können nicht storniert werden"}
	}

	settings, err := store.GetSettings(ctx, db)
	if err != nil {
		return fmt.Errorf("kiosk cancel: get settings: %w", err)
	}

	if time.Since(txn.CreatedAt) > time.Duration(settings.CancellationMinutes)*time.Minute {
		return &handler.ValidationError{Message: "Stornierungsfenster abgelaufen"}
	}

	err = h.Store.WithTx(ctx, func(tx pgx.Tx) error {
		return store.CancelTransaction(ctx, tx, id, user.ID)
	})
	if err != nil {
		return fmt.Errorf("kiosk cancel: %w", err)
	}

	// Re-render the kiosk history list.
	txns, err := store.ListTransactionsByCreator(ctx, db, user.ID, 20)
	if err != nil {
		return fmt.Errorf("kiosk cancel: list transactions: %w", err)
	}

	// Render toast + refreshed history list + close modal.
	h.Renderer.Fragment(w, r, "toast", map[string]string{
		"Type":    "success",
		"Message": "Transaktion storniert!",
	})

	h.Renderer.AppendOOB(w, "kiosk-history-list", map[string]any{
		"Transactions": txns,
		"Settings":     settings,
		"KioskUserID":  user.ID,
		"OOB":          true,
	})

	w.Write([]byte(`<div id="modal" hx-swap-oob="innerHTML" style="display:none"></div>`))

	return nil
}
