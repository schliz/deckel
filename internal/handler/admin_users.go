package handler

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/k4-bar/deckel/internal/auth"
	"github.com/k4-bar/deckel/internal/middleware"
	"github.com/k4-bar/deckel/internal/model"
	"github.com/k4-bar/deckel/internal/store"
	"github.com/jackc/pgx/v5"
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

	h.Renderer.Page(w, r, "admin_users", data)
	return nil
}

// ToggleActive handles POST /admin/users/{id}/toggle-active.
func (h *Handler) ToggleActive(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	db := h.Store.DB()

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		return &ValidationError{Message: "ungültige Benutzer-ID"}
	}

	if err := store.ToggleActive(ctx, db, id); err != nil {
		return fmt.Errorf("toggle active: %w", err)
	}

	ub, err := store.GetUserWithBalance(ctx, db, id)
	if err != nil {
		return fmt.Errorf("toggle active: fetch user: %w", err)
	}

	h.Renderer.Fragment(w, r, "user-row", *ub)
	h.Renderer.AppendOOB(w, "toast", map[string]string{
		"Type":    "success",
		"Message": fmt.Sprintf("Status für %s geändert.", ub.FullName),
	})
	return nil
}

// ToggleSpendingLimit handles POST /admin/users/{id}/toggle-spending-limit.
func (h *Handler) ToggleSpendingLimit(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	db := h.Store.DB()

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		return &ValidationError{Message: "ungültige Benutzer-ID"}
	}

	if err := store.ToggleSpendingLimit(ctx, db, id); err != nil {
		return fmt.Errorf("toggle spending limit: %w", err)
	}

	ub, err := store.GetUserWithBalance(ctx, db, id)
	if err != nil {
		return fmt.Errorf("toggle spending limit: fetch user: %w", err)
	}

	h.Renderer.Fragment(w, r, "user-row", *ub)
	h.Renderer.AppendOOB(w, "toast", map[string]string{
		"Type":    "success",
		"Message": fmt.Sprintf("Ausgabelimit für %s geändert.", ub.FullName),
	})
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

	data := map[string]any{
		"User":      user,
		"CSRFToken": middleware.CSRFTokenFromContext(ctx),
	}

	h.Renderer.Fragment(w, r, "payment-modal", data)
	return nil
}

// ToggleBarteamer handles POST /admin/users/{id}/toggle-barteamer.
func (h *Handler) ToggleBarteamer(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	db := h.Store.DB()

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		return &ValidationError{Message: "ungültige Benutzer-ID"}
	}

	if err := store.ToggleBarteamer(ctx, db, id); err != nil {
		return fmt.Errorf("toggle barteamer: %w", err)
	}

	ub, err := store.GetUserWithBalance(ctx, db, id)
	if err != nil {
		return fmt.Errorf("toggle barteamer: fetch user: %w", err)
	}

	h.Renderer.Fragment(w, r, "user-row", *ub)
	h.Renderer.AppendOOB(w, "toast", map[string]string{
		"Type":    "success",
		"Message": fmt.Sprintf("Status für %s geändert.", ub.FullName),
	})
	return nil
}

// RegisterDeposit handles POST /admin/users/{id}/deposit.
func (h *Handler) RegisterDeposit(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	db := h.Store.DB()

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		return &ValidationError{Message: "ungültige Benutzer-ID"}
	}

	// Parse Euro amount string to cents.
	amountStr := r.FormValue("amount")
	amountFloat, err := strconv.ParseFloat(amountStr, 64)
	if err != nil || amountFloat <= 0 {
		return &ValidationError{Message: "Ungültiger Betrag"}
	}
	amountCents := int64(math.Round(amountFloat * 100))

	// Optional note, default to "Einzahlung".
	note := strings.TrimSpace(r.FormValue("note"))
	if note == "" {
		note = "Einzahlung"
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

	// Build response: toast + OOB header-stats + OOB user row + close modal.
	h.Renderer.Fragment(w, r, "toast", map[string]string{
		"Type":    "success",
		"Message": "Einzahlung gebucht!",
	})

	// Fetch settings for header-stats.
	settings, err := store.GetSettings(ctx, db)
	if err != nil {
		return fmt.Errorf("register deposit: get settings: %w", err)
	}

	// Render OOB header-stats update.
	reqUser := auth.UserFromContext(ctx)
	newBalance, _ := store.GetUserBalance(ctx, db, reqUser.ID)
	totalBalance, _ := store.GetAllBalancesSum(ctx, db)
	rank, total, _ := store.GetUserRank(ctx, db, reqUser.ID)

	h.Renderer.AppendOOB(w, "header-stats", map[string]any{
		"UserBalance":  newBalance,
		"TotalBalance": totalBalance,
		"UserRank":     rank,
		"TotalUsers":   total,
		"Settings":     settings,
		"OOB":          true,
	})

	// Render OOB user row swap.
	ub, err := store.GetUserWithBalance(ctx, db, id)
	if err != nil {
		return fmt.Errorf("register deposit: fetch user: %w", err)
	}
	h.Renderer.AppendOOB(w, "user-row", *ub)

	// Close modal.
	w.Write([]byte(`<div id="modal" hx-swap-oob="innerHTML" style="display:none"></div>`))

	return nil
}
