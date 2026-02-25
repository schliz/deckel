package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/k4-bar/deckel/internal/auth"
	"github.com/k4-bar/deckel/internal/middleware"
	"github.com/k4-bar/deckel/internal/model"
	"github.com/k4-bar/deckel/internal/store"
)

// AdminUsersPageData is the view model for the admin user list page.
type AdminUsersPageData struct {
	User       *auth.RequestUser
	Users      []model.UserWithBalance
	Settings   *model.Settings
	CSRFToken  string
	ActivePage string
	Page       int
	TotalPages int
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
		User:       user,
		Users:      users,
		Settings:   settings,
		CSRFToken:  middleware.CSRFTokenFromContext(ctx),
		ActivePage: "admin-users",
		Page:       page,
		TotalPages: totalPages,
	}

	h.Renderer.Page(w, r, "admin_users", data)
	return nil
}
