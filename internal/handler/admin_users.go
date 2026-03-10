package handler

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/schliz/deckel/internal/auth"
	"github.com/schliz/deckel/internal/middleware"
	"github.com/schliz/deckel/internal/model"
	"github.com/schliz/deckel/internal/store"
)

// AdminUsersPageData is the view model for the admin user list page.
type AdminUsersPageData struct {
	User              *auth.RequestUser
	Users             []model.UserWithBalance
	Settings          *model.Settings
	CSRFToken         string
	ActivePage        string
	Page              int
	TotalPages        int
	LowBalanceWarning bool
}

// rejectKioskTarget returns a ForbiddenError if the target user is a kiosk account.
func rejectKioskTarget(ctx context.Context, db store.DBTX, id int64) error {
	u, err := store.GetUserWithBalance(ctx, db, id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return &NotFoundError{Message: "Benutzer nicht gefunden"}
		}
		return fmt.Errorf("check kiosk target: %w", err)
	}
	if u.IsKiosk {
		return &ForbiddenError{Message: "Diese Aktion ist für Kiosk-Benutzer nicht verfügbar."}
	}
	return nil
}

// userRowData builds the template data map for a user-row fragment.
func userRowData(ub model.UserWithBalance, currentUserID int64) map[string]any {
	return map[string]any{
		"User":          ub,
		"CurrentUserID": currentUserID,
	}
}

// AdminUserList renders the paginated admin user list sorted by balance ascending.
func (h *Handler) AdminUserList(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	user := auth.UserFromContext(ctx)
	db := h.Store.DB()

	// Fetch settings for pagination size.
	settings, err := store.GetSettings(ctx, db)
	if err != nil {
		return fmt.Errorf("admin user list: get settings: %w", err)
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

	// Fetch paginated users sorted by balance ascending.
	users, total, err := store.ListUsersWithBalance(ctx, db, limit, offset)
	if err != nil {
		return fmt.Errorf("admin user list: list users: %w", err)
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

	data := AdminUsersPageData{
		User:              user,
		Users:             users,
		Settings:          settings,
		CSRFToken:         middleware.CSRFTokenFromContext(ctx),
		ActivePage:        "admin-users",
		Page:              page,
		TotalPages:        totalPages,
		LowBalanceWarning: isLowBalance(user, settings),
	}

	if isHTMX(r) {
		h.Renderer.Fragment(w, r, "user-list", data)
		return nil
	}

	h.Renderer.Page(w, r, "admin_users", data)
	return nil
}

// ToggleActive handles POST /admin/users/{id}/toggle-active.
func (h *Handler) ToggleActive(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	reqUser := auth.UserFromContext(ctx)
	db := h.Store.DB()

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		return &ValidationError{Message: "ungültige Benutzer-ID"}
	}

	if id == reqUser.ID {
		return &ForbiddenError{Message: "Du kannst deinen eigenen Account nicht deaktivieren."}
	}

	if err := store.ToggleActive(ctx, db, id); err != nil {
		return fmt.Errorf("toggle active: %w", err)
	}

	ub, err := store.GetUserWithBalance(ctx, db, id)
	if err != nil {
		return fmt.Errorf("toggle active: fetch user: %w", err)
	}

	h.Renderer.Fragment(w, r, "user-row", userRowData(*ub, reqUser.ID))
	h.Renderer.AppendOOB(w, "toast", map[string]string{
		"Type":    "success",
		"Message": fmt.Sprintf("Status für %s geändert.", ub.FullName),
	})
	return nil
}

// ToggleSpendingLimit handles POST /admin/users/{id}/toggle-spending-limit.
func (h *Handler) ToggleSpendingLimit(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	reqUser := auth.UserFromContext(ctx)
	db := h.Store.DB()

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		return &ValidationError{Message: "ungültige Benutzer-ID"}
	}

	if err := rejectKioskTarget(ctx, db, id); err != nil {
		return err
	}

	if err := store.ToggleSpendingLimit(ctx, db, id); err != nil {
		return fmt.Errorf("toggle spending limit: %w", err)
	}

	ub, err := store.GetUserWithBalance(ctx, db, id)
	if err != nil {
		return fmt.Errorf("toggle spending limit: fetch user: %w", err)
	}

	h.Renderer.Fragment(w, r, "user-row", userRowData(*ub, reqUser.ID))
	h.Renderer.AppendOOB(w, "toast", map[string]string{
		"Type":    "success",
		"Message": fmt.Sprintf("Ausgabelimit für %s geändert.", ub.FullName),
	})
	return nil
}

// ToggleBarteamer handles POST /admin/users/{id}/toggle-barteamer.
func (h *Handler) ToggleBarteamer(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	reqUser := auth.UserFromContext(ctx)
	db := h.Store.DB()

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		return &ValidationError{Message: "ungültige Benutzer-ID"}
	}

	if err := rejectKioskTarget(ctx, db, id); err != nil {
		return err
	}

	if err := store.ToggleBarteamer(ctx, db, id); err != nil {
		return fmt.Errorf("toggle barteamer: %w", err)
	}

	ub, err := store.GetUserWithBalance(ctx, db, id)
	if err != nil {
		return fmt.Errorf("toggle barteamer: fetch user: %w", err)
	}

	h.Renderer.Fragment(w, r, "user-row", userRowData(*ub, reqUser.ID))
	h.Renderer.AppendOOB(w, "toast", map[string]string{
		"Type":    "success",
		"Message": fmt.Sprintf("Status für %s geändert.", ub.FullName),
	})
	return nil
}

// ConfirmToggleModal handles GET /admin/users/{id}/confirm-toggle?action=...
// and returns a confirmation modal fragment.
func (h *Handler) ConfirmToggleModal(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	reqUser := auth.UserFromContext(ctx)
	db := h.Store.DB()

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		return &ValidationError{Message: "ungültige Benutzer-ID"}
	}

	action := r.URL.Query().Get("action")

	ub, err := store.GetUserWithBalance(ctx, db, id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return &NotFoundError{Message: "Benutzer nicht gefunden"}
		}
		return fmt.Errorf("confirm toggle modal: get user: %w", err)
	}

	var title, message, postURL string
	isSelf := id == reqUser.ID

	// Block barteamer/spending-limit/deposit actions for kiosk users.
	if ub.IsKiosk && action != "active" {
		return &ForbiddenError{Message: "Diese Aktion ist für Kiosk-Benutzer nicht verfügbar."}
	}

	switch action {
	case "active":
		postURL = fmt.Sprintf("/admin/users/%d/toggle-active", id)
		if ub.IsActive {
			title = "Benutzer deaktivieren"
			message = fmt.Sprintf("Möchtest du %s wirklich deaktivieren? Der Benutzer kann sich dann nicht mehr anmelden.", ub.FullName)
		} else {
			title = "Benutzer aktivieren"
			message = fmt.Sprintf("Möchtest du %s wieder aktivieren?", ub.FullName)
		}
		if isSelf {
			message = "Du kannst deinen eigenen Account nicht deaktivieren."
		}
	case "barteamer":
		postURL = fmt.Sprintf("/admin/users/%d/toggle-barteamer", id)
		if ub.IsBarteamer {
			title = "Auf Helfer setzen"
			message = fmt.Sprintf("Möchtest du %s auf Helfer-Preise umstellen?", ub.FullName)
		} else {
			title = "Auf Barteamer setzen"
			message = fmt.Sprintf("Möchtest du %s auf Barteamer-Preise umstellen?", ub.FullName)
		}
	case "spending-limit":
		postURL = fmt.Sprintf("/admin/users/%d/toggle-spending-limit", id)
		if ub.SpendingLimitDisabled {
			title = "Ausgabelimit aktivieren"
			message = fmt.Sprintf("Möchtest du das Ausgabelimit für %s wieder aktivieren?", ub.FullName)
		} else {
			title = "Ausgabelimit aufheben"
			message = fmt.Sprintf("Möchtest du das Ausgabelimit für %s aufheben?", ub.FullName)
		}
	default:
		return &ValidationError{Message: "ungültige Aktion"}
	}

	data := map[string]any{
		"Title":     title,
		"Message":   message,
		"PostURL":   postURL,
		"UserID":    id,
		"IsSelf":    isSelf && action == "active",
		"CSRFToken": middleware.CSRFTokenFromContext(ctx),
	}

	h.Renderer.Fragment(w, r, "confirm-toggle-modal", data)
	return nil
}

// DepositModal handles GET /admin/users/{id}/deposit.
func (h *Handler) DepositModal(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	db := h.Store.DB()

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		return &ValidationError{Message: "ungültige Benutzer-ID"}
	}

	user, err := store.GetUserWithBalance(ctx, db, id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return &NotFoundError{Message: "Benutzer nicht gefunden"}
		}
		return fmt.Errorf("deposit modal: get user: %w", err)
	}
	if user.IsKiosk {
		return &ForbiddenError{Message: "Diese Aktion ist für Kiosk-Benutzer nicht verfügbar."}
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
		return &ValidationError{Message: "ungültige Benutzer-ID"}
	}

	if err := rejectKioskTarget(ctx, db, id); err != nil {
		return err
	}

	// Parse Euro amount string to cents (accepts both comma and period).
	amountFloat, err := strconv.ParseFloat(normalizeDecimal(r.FormValue("amount")), 64)
	if err != nil || amountFloat <= 0 {
		return &ValidationError{Message: "Ungültiger Betrag"}
	}
	amountCents := int64(math.Round(amountFloat * 100))

	// Optional note, default to "Einzahlung".
	note := strings.TrimSpace(r.FormValue("note"))
	if note == "" {
		note = "Einzahlung"
	}
	if err := validateTextLen(note, 500, "Notiz"); err != nil {
		return err
	}

	// Create deposit transaction (positive amount = credit).
	txn := &model.Transaction{
		UserID:      id,
		Amount:      amountCents,
		Description: &note,
		Type:        "deposit",
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
