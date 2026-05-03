package kiosk

// Kiosk transaction history view: lists recent transactions created by the
// currently authenticated kiosk user.

import (
	"fmt"
	"net/http"

	"github.com/schliz/deckel/internal/auth"
	"github.com/schliz/deckel/internal/middleware"
	"github.com/schliz/deckel/internal/store"
)

// KioskHistory renders the recent kiosk transaction history.
func (h *Handler) KioskHistory(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	user := auth.UserFromContext(ctx)
	db := h.Store.DB()

	txns, err := store.ListTransactionsByCreator(ctx, db, user.ID, 20)
	if err != nil {
		return fmt.Errorf("kiosk history: list transactions: %w", err)
	}

	settings, err := store.GetSettings(ctx, db)
	if err != nil {
		return fmt.Errorf("kiosk history: get settings: %w", err)
	}

	data := map[string]any{
		"Transactions": txns,
		"Settings":     settings,
		"CSRFToken":    middleware.CSRFTokenFromContext(ctx),
		"KioskUserID":  user.ID,
	}

	h.Renderer.Page(w, r, "kiosk/history", data)
	return nil
}
