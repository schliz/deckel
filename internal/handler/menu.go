package handler

import (
	"fmt"
	"net/http"

	"github.com/k4-bar/deckel/internal/auth"
	"github.com/k4-bar/deckel/internal/middleware"
	"github.com/k4-bar/deckel/internal/model"
	"github.com/k4-bar/deckel/internal/store"
)

// CategoryWithItems groups a category with its active items.
type CategoryWithItems struct {
	Category model.Category
	Items    []model.Item
}

// MenuPageData is the view model for the menu page.
type MenuPageData struct {
	User       *auth.RequestUser
	Categories []CategoryWithItems
	Settings   *model.Settings
	CSRFToken  string
	IsBlocked  bool
	ActivePage string
}

// MenuPage renders the drinks menu organized by category.
func (h *Handler) MenuPage(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	user := auth.UserFromContext(ctx)

	db := h.Store.DB()

	// Fetch settings.
	settings, err := store.GetSettings(ctx, db)
	if err != nil {
		return fmt.Errorf("menu: get settings: %w", err)
	}

	// Fetch categories.
	cats, err := store.ListCategories(ctx, db)
	if err != nil {
		return fmt.Errorf("menu: list categories: %w", err)
	}

	// Fetch items for each category.
	var categories []CategoryWithItems
	for _, cat := range cats {
		items, err := store.ListItemsByCategory(ctx, db, cat.ID)
		if err != nil {
			return fmt.Errorf("menu: list items for category %d: %w", cat.ID, err)
		}
		if len(items) > 0 {
			categories = append(categories, CategoryWithItems{
				Category: cat,
				Items:    items,
			})
		}
	}

	// Determine if user is blocked (at or below hard spending limit).
	isBlocked := false
	if settings.HardLimitEnabled && user != nil {
		isBlocked = user.Balance <= -settings.HardSpendingLimit
	}

	data := MenuPageData{
		User:       user,
		Categories: categories,
		Settings:   settings,
		CSRFToken:  middleware.CSRFTokenFromContext(ctx),
		IsBlocked:  isBlocked,
		ActivePage: "menu",
	}

	h.Renderer.Page(w, r, "menu", data)
	return nil
}
