package admin

// Admin deposit flow: modal that prompts for an amount, plus the handler
// that books the deposit transaction and refreshes the user row and header
// stats. Uses the rejectKioskTarget and userRowData helpers from users.go.

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/schliz/deckel/internal/auth"
	"github.com/schliz/deckel/internal/handler"
	"github.com/schliz/deckel/internal/middleware"
	"github.com/schliz/deckel/internal/model"
	"github.com/schliz/deckel/internal/store"
)

// DepositModal handles GET /admin/users/{id}/deposit.
func (h *Handler) DepositModal(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	db := h.Store.DB()

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		return &handler.ValidationError{Message: "ungültige Benutzer-ID"}
	}

	user, err := store.GetUserWithBalance(ctx, db, id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return &handler.NotFoundError{Message: "Benutzer nicht gefunden"}
		}
		return fmt.Errorf("deposit modal: get user: %w", err)
	}
	if user.IsKiosk {
		return &handler.ForbiddenError{Message: "Diese Aktion ist für Kiosk-Benutzer nicht verfügbar."}
	}

	data := map[string]any{
		"User":      user,
		"CSRFToken": middleware.CSRFTokenFromContext(ctx),
	}

	h.Renderer.Fragment(w, r, "payment-modal", data)
	return nil
}

// RegisterDeposit handles POST /admin/users/{id}/deposit.
func (h *Handler) RegisterDeposit(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	reqUser := auth.UserFromContext(ctx)
	db := h.Store.DB()

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		return &handler.ValidationError{Message: "ungültige Benutzer-ID"}
	}

	if err := rejectKioskTarget(ctx, db, id); err != nil {
		return err
	}

	// Parse Euro amount string to cents (accepts both comma and period).
	amountFloat, err := strconv.ParseFloat(handler.NormalizeDecimal(r.FormValue("amount")), 64)
	if err != nil || amountFloat <= 0 {
		return &handler.ValidationError{Message: "Ungültiger Betrag"}
	}
	amountCents := int64(math.Round(amountFloat * 100))

	// Optional note, default to "Einzahlung".
	note := strings.TrimSpace(r.FormValue("note"))
	if note == "" {
		note = "Einzahlung"
	}
	if err := handler.ValidateTextLen(note, 500, "Notiz"); err != nil {
		return err
	}

	// Create deposit transaction (positive amount = credit).
	createdBy := reqUser.ID
	txn := &model.Transaction{
		UserID:          id,
		Amount:          amountCents,
		Description:     &note,
		Type:            "deposit",
		CreatedByUserID: &createdBy,
	}

	err = h.Store.WithTx(ctx, func(tx pgx.Tx) error {
		_, err := store.CreateTransaction(ctx, tx, txn)
		return err
	})
	if err != nil {
		return fmt.Errorf("register deposit: %w", err)
	}

	// Build response: user-row as primary body + OOB toast + OOB header-stats.
	// Modal is closed client-side via hx-on::after-request on the button.

	// Render user row as primary body (swapped into hx-target via outerHTML).
	ub, err := store.GetUserWithBalance(ctx, db, id)
	if err != nil {
		return fmt.Errorf("register deposit: fetch user: %w", err)
	}
	h.Renderer.Fragment(w, r, "user-row", userRowData(*ub, reqUser.ID))

	// OOB toast.
	h.Renderer.AppendOOB(w, "toast", map[string]string{
		"Type":    "success",
		"Message": "Einzahlung gebucht!",
	})

	// Fetch settings for header-stats.
	settings, err := store.GetSettings(ctx, db)
	if err != nil {
		return fmt.Errorf("register deposit: get settings: %w", err)
	}

	// OOB header-stats update.
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

	return nil
}
