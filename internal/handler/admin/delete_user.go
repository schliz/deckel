package admin

// Admin user deletion: confirmation modal with guard checks (self, active
// state, balance, remaining transaction references) plus the delete action
// itself, which re-checks the same guards under a row lock inside a
// transaction to close the race against concurrent bookings.

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5"

	"github.com/schliz/deckel/internal/auth"
	"github.com/schliz/deckel/internal/handler"
	"github.com/schliz/deckel/internal/middleware"
	"github.com/schliz/deckel/internal/store"
)

// ConfirmDeleteUserModal handles GET /admin/users/{id}/confirm-delete and
// returns a confirmation modal fragment. If any guard fails, the modal shows
// the reason instead of a confirm button (blocked state, same pattern as
// confirm-toggle-modal's IsSelf case).
func (h *Handler) ConfirmDeleteUserModal(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	reqUser := auth.UserFromContext(ctx)
	db := h.Store.DB()

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		return &handler.ValidationError{Message: "ungültige Benutzer-ID"}
	}

	ub, err := store.GetUserWithBalance(ctx, db, id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return &handler.NotFoundError{Message: "Benutzer nicht gefunden"}
		}
		return fmt.Errorf("confirm delete user modal: get user: %w", err)
	}

	blocked := ""
	switch {
	case id == reqUser.ID:
		blocked = "Du kannst deinen eigenen Account nicht löschen."
	case ub.IsActive:
		blocked = "Benutzer muss zuerst deaktiviert werden."
	case ub.Balance != 0:
		blocked = "Kontostand muss 0,00 € betragen."
	default:
		asUser, asCreator, err := store.CountUserReferences(ctx, db, id)
		if err != nil {
			return fmt.Errorf("confirm delete user modal: count references: %w", err)
		}
		if asUser > 0 || asCreator > 0 {
			blocked = "Es existieren noch zugeordnete Buchungen — bitte auf die Anonymisierung warten."
		}
	}

	data := map[string]any{
		"Title":     "Benutzer löschen",
		"Message":   fmt.Sprintf("Möchtest du %s endgültig löschen? Dies kann nicht rückgängig gemacht werden.", ub.FullName),
		"PostURL":   fmt.Sprintf("/admin/users/%d/delete", id),
		"UserID":    id,
		"Blocked":   blocked,
		"CSRFToken": middleware.CSRFTokenFromContext(ctx),
	}

	h.Renderer.Fragment(w, r, "confirm-delete-user-modal", data)
	return nil
}

// DeleteUser handles POST /admin/users/{id}/delete. All guards are
// re-checked inside the transaction with a row lock (FOR UPDATE) to protect
// against a booking landing concurrently with the delete.
func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	reqUser := auth.UserFromContext(ctx)

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		return &handler.ValidationError{Message: "ungültige Benutzer-ID"}
	}

	if id == reqUser.ID {
		return &handler.ForbiddenError{Message: "Du kannst deinen eigenen Account nicht löschen."}
	}

	err = h.Store.WithTx(ctx, func(tx pgx.Tx) error {
		var isActive bool
		var balance int64
		err := tx.QueryRow(ctx, `
			SELECT u.is_active, u.balance_carryover + COALESCE((
				SELECT SUM(amount) FROM transactions WHERE user_id = u.id
			), 0)
			FROM users u WHERE u.id = $1 FOR UPDATE`, id).Scan(&isActive, &balance)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return store.ErrNotFound
			}
			return fmt.Errorf("delete user: lock user: %w", err)
		}

		if isActive {
			return &handler.ValidationError{Message: "Benutzer muss zuerst deaktiviert werden"}
		}
		if balance != 0 {
			return &handler.ValidationError{Message: "Kontostand muss 0,00 € betragen"}
		}

		asUser, asCreator, err := store.CountUserReferences(ctx, tx, id)
		if err != nil {
			return fmt.Errorf("delete user: count references: %w", err)
		}
		if asUser > 0 || asCreator > 0 {
			return &handler.ValidationError{Message: "Es existieren noch zugeordnete Buchungen — bitte auf die Anonymisierung warten"}
		}

		if err := store.DeleteUser(ctx, tx, id); err != nil {
			return fmt.Errorf("delete user: %w", err)
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return &handler.NotFoundError{Message: "Benutzer nicht gefunden"}
		}
		return err
	}

	// The confirm button targets #user-row-{id} with hx-swap="delete", which
	// removes the row on any 2xx response regardless of body content — no
	// primary fragment needed. Never hx-swap-oob a <tr> (see CLAUDE.md).
	h.Renderer.AppendOOB(w, "toast", map[string]string{
		"Type":    "success",
		"Message": "Benutzer gelöscht",
	})
	return nil
}
