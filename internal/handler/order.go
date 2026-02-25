package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/k4-bar/deckel/internal/auth"
	"github.com/k4-bar/deckel/internal/middleware"
	"github.com/k4-bar/deckel/internal/model"
	"github.com/k4-bar/deckel/internal/store"
)

// OrderModalData is the view model for the order confirmation modal.
type OrderModalData struct {
	Item        *model.Item
	User        *auth.RequestUser
	MaxQuantity int
	CSRFToken   string
}

// OrderModal renders the order confirmation modal for a given item.
func (h *Handler) OrderModal(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	user := auth.UserFromContext(ctx)
	db := h.Store.DB()

	// Extract item ID from URL path.
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return &NotFoundError{Message: "Artikel nicht gefunden"}
	}

	// Fetch item from DB.
	item, err := store.GetItem(ctx, db, id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return &NotFoundError{Message: "Artikel nicht gefunden"}
		}
		return fmt.Errorf("order modal: get item: %w", err)
	}

	// Return 404 if soft-deleted.
	if item.DeletedAt != nil {
		return &NotFoundError{Message: "Artikel nicht gefunden"}
	}

	// Fetch settings for max_item_quantity.
	settings, err := store.GetSettings(ctx, db)
	if err != nil {
		return fmt.Errorf("order modal: get settings: %w", err)
	}

	data := OrderModalData{
		Item:        item,
		User:        user,
		MaxQuantity: settings.MaxItemQuantity,
		CSRFToken:   middleware.CSRFTokenFromContext(ctx),
	}

	h.Renderer.Fragment(w, r, "order-modal", data)
	return nil
}
