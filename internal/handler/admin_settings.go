package handler

import (
	"fmt"
	"net/http"

	"github.com/k4-bar/deckel/internal/auth"
	"github.com/k4-bar/deckel/internal/middleware"
	"github.com/k4-bar/deckel/internal/model"
	"github.com/k4-bar/deckel/internal/store"
)

// AdminSettingsPageData is the view model for the admin settings page.
type AdminSettingsPageData struct {
	User       *auth.RequestUser
	Settings   *model.Settings
	CSRFToken  string
	ActivePage string
}

// AdminSettingsPage renders the admin settings page.
func (h *Handler) AdminSettingsPage(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	user := auth.UserFromContext(ctx)
	db := h.Store.DB()

	settings, err := store.GetSettings(ctx, db)
	if err != nil {
		return fmt.Errorf("admin settings: get settings: %w", err)
	}

	data := AdminSettingsPageData{
		User:       user,
		Settings:   settings,
		CSRFToken:  middleware.CSRFTokenFromContext(ctx),
		ActivePage: "admin-settings",
	}

	h.Renderer.Page(w, r, "admin_settings", data)
	return nil
}
