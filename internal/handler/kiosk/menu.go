package kiosk

// Kiosk pre-order UI: pick an item from the menu grid, then pick the user the
// purchase will be booked against.

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/schliz/deckel/internal/auth"
	"github.com/schliz/deckel/internal/handler"
	"github.com/schliz/deckel/internal/middleware"
	"github.com/schliz/deckel/internal/model"
	"github.com/schliz/deckel/internal/store"
)

// KioskMenuPage renders the kiosk item selection grid.
func (h *Handler) KioskMenuPage(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	user := auth.UserFromContext(ctx)
	db := h.Store.DB()

	cats, err := store.ListCategories(ctx, db)
	if err != nil {
		return fmt.Errorf("kiosk menu: list categories: %w", err)
	}

	var categories []handler.CategoryWithItems
	for _, cat := range cats {
		items, err := store.ListItemsByCategory(ctx, db, cat.ID)
		if err != nil {
			return fmt.Errorf("kiosk menu: list items for category %d: %w", cat.ID, err)
		}
		if len(items) > 0 {
			categories = append(categories, handler.CategoryWithItems{
				Category: cat,
				Items:    items,
			})
		}
	}

	data := map[string]any{
		"Categories": categories,
		"CSRFToken":  middleware.CSRFTokenFromContext(ctx),
		"User":       user,
	}

	h.Renderer.Page(w, r, "kiosk/menu", data)
	return nil
}

// KioskUserSelect renders the user selection page for a kiosk order.
func (h *Handler) KioskUserSelect(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	user := auth.UserFromContext(ctx)
	db := h.Store.DB()

	idStr := r.PathValue("id")
	itemID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return &handler.NotFoundError{Message: "Artikel nicht gefunden"}
	}

	item, err := store.GetItem(ctx, db, itemID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return &handler.NotFoundError{Message: "Artikel nicht gefunden"}
		}
		return fmt.Errorf("kiosk user select: get item: %w", err)
	}
	if item.DeletedAt != nil {
		return &handler.NotFoundError{Message: "Artikel nicht gefunden"}
	}

	users, err := store.ListActiveUsersWithBalance(ctx, db)
	if err != nil {
		return fmt.Errorf("kiosk user select: list users: %w", err)
	}

	// Filter out the kiosk user itself.
	var filtered []model.UserWithBalance
	for _, u := range users {
		if u.ID != user.ID {
			filtered = append(filtered, u)
		}
	}

	settings, err := store.GetSettings(ctx, db)
	if err != nil {
		return fmt.Errorf("kiosk user select: get settings: %w", err)
	}

	data := map[string]any{
		"Item":      item,
		"Users":     filtered,
		"Settings":  settings,
		"CSRFToken": middleware.CSRFTokenFromContext(ctx),
	}

	h.Renderer.Page(w, r, "kiosk/users", data)
	return nil
}
