package handler

import (
	"fmt"
	"net/http"

	"github.com/k4-bar/deckel/internal/auth"
	"github.com/k4-bar/deckel/internal/middleware"
	"github.com/k4-bar/deckel/internal/model"
	"github.com/k4-bar/deckel/internal/store"
)

// AdminMenuPageData is the view model for the admin menu management page.
type AdminMenuPageData struct {
	User       *auth.RequestUser
	Categories []CategoryWithItems
	Settings   *model.Settings
	CSRFToken  string
	ActivePage string
}

// AdminMenuPage renders the admin menu management page with all categories and items.
func (h *Handler) AdminMenuPage(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	user := auth.UserFromContext(ctx)

	db := h.Store.DB()

	// Fetch settings.
	settings, err := store.GetSettings(ctx, db)
	if err != nil {
		return fmt.Errorf("admin menu: get settings: %w", err)
	}

	// Fetch all categories.
	cats, err := store.ListCategories(ctx, db)
	if err != nil {
		return fmt.Errorf("admin menu: list categories: %w", err)
	}

	// Fetch items for each category (including empty categories for admin view).
	var categories []CategoryWithItems
	for _, cat := range cats {
		items, err := store.ListItemsByCategory(ctx, db, cat.ID)
		if err != nil {
			return fmt.Errorf("admin menu: list items for category %d: %w", cat.ID, err)
		}
		categories = append(categories, CategoryWithItems{
			Category: cat,
			Items:    items,
		})
	}

	data := AdminMenuPageData{
		User:       user,
		Categories: categories,
		Settings:   settings,
		CSRFToken:  middleware.CSRFTokenFromContext(ctx),
		ActivePage: "admin-menu",
	}

	h.Renderer.Page(w, r, "admin_menu", data)
	return nil
}
