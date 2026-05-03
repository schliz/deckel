package admin

// Admin menu management: item-level handlers (create, edit, reorder, soft
// delete). Each mutation re-renders the category list via the helper in
// categories.go.

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/schliz/deckel/internal/handler"
	"github.com/schliz/deckel/internal/middleware"
	"github.com/schliz/deckel/internal/model"
	"github.com/schliz/deckel/internal/store"
)

// CreateItem handles POST /admin/categories/{cat_id}/items to add a new drink item.
func (h *Handler) CreateItem(w http.ResponseWriter, r *http.Request) error {
	catIDStr := r.PathValue("id")
	catID, err := strconv.ParseInt(catIDStr, 10, 64)
	if err != nil {
		return &handler.NotFoundError{Message: "Kategorie nicht gefunden"}
	}

	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		return &handler.ValidationError{Message: "Artikelname darf nicht leer sein"}
	}

	if err := handler.ValidateTextLen(name, 255, "Artikelname"); err != nil {
		return err
	}

	priceBarteamerF, err := strconv.ParseFloat(handler.NormalizeDecimal(r.FormValue("price_barteamer")), 64)
	if err != nil || priceBarteamerF <= 0 {
		return &handler.ValidationError{Message: "Barteamer-Preis muss größer als 0 sein"}
	}
	priceHelferF, err := strconv.ParseFloat(handler.NormalizeDecimal(r.FormValue("price_helfer")), 64)
	if err != nil || priceHelferF <= 0 {
		return &handler.ValidationError{Message: "Helfer-Preis muss größer als 0 sein"}
	}

	priceBarteamer := int64(math.Round(priceBarteamerF * 100))
	priceHelfer := int64(math.Round(priceHelferF * 100))

	ctx := r.Context()
	db := h.Store.DB()

	_, err = store.CreateItem(ctx, db, &model.Item{
		CategoryID:     catID,
		Name:           name,
		PriceBarteamer: priceBarteamer,
		PriceHelfer:    priceHelfer,
	})
	if err != nil {
		return fmt.Errorf("create item: %w", err)
	}

	if err := h.renderAdminCategoryList(w, r); err != nil {
		return err
	}

	h.Renderer.AppendOOB(w, "toast", map[string]string{
		"Type":    "success",
		"Message": fmt.Sprintf("Artikel '%s' erstellt", name),
	})

	return nil
}

// ReorderItem handles POST /admin/items/{id}/reorder to change item order within its category.
func (h *Handler) ReorderItem(w http.ResponseWriter, r *http.Request) error {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return &handler.NotFoundError{Message: "Artikel nicht gefunden"}
	}

	direction := r.URL.Query().Get("direction")
	if direction == "" {
		direction = r.FormValue("direction")
	}

	ctx := r.Context()
	db := h.Store.DB()

	if err := store.ReorderItem(ctx, db, id, direction); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return &handler.NotFoundError{Message: "Artikel nicht gefunden"}
		}
		return fmt.Errorf("reorder item: %w", err)
	}

	return h.renderAdminCategoryList(w, r)
}

// EditItemForm handles GET /admin/items/{id}/edit and returns an edit modal fragment.
func (h *Handler) EditItemForm(w http.ResponseWriter, r *http.Request) error {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return &handler.NotFoundError{Message: "Artikel nicht gefunden"}
	}

	ctx := r.Context()
	db := h.Store.DB()

	item, err := store.GetItem(ctx, db, id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return &handler.NotFoundError{Message: "Artikel nicht gefunden"}
		}
		return fmt.Errorf("edit item form: %w", err)
	}
	if item.DeletedAt != nil {
		return &handler.NotFoundError{Message: "Artikel nicht gefunden"}
	}

	data := struct {
		Item      *model.Item
		CSRFToken string
	}{
		Item:      item,
		CSRFToken: middleware.CSRFTokenFromContext(ctx),
	}

	h.Renderer.Fragment(w, r, "edit-item-modal", data)
	return nil
}

// UpdateItem handles POST /admin/items/{id}/update to modify a drink item.
func (h *Handler) UpdateItem(w http.ResponseWriter, r *http.Request) error {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return &handler.NotFoundError{Message: "Artikel nicht gefunden"}
	}

	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		return &handler.ValidationError{Message: "Artikelname darf nicht leer sein"}
	}
	if err := handler.ValidateTextLen(name, 255, "Artikelname"); err != nil {
		return err
	}

	priceBarteamerF, err := strconv.ParseFloat(handler.NormalizeDecimal(r.FormValue("price_barteamer")), 64)
	if err != nil || priceBarteamerF <= 0 {
		return &handler.ValidationError{Message: "Barteamer-Preis muss größer als 0 sein"}
	}
	priceHelferF, err := strconv.ParseFloat(handler.NormalizeDecimal(r.FormValue("price_helfer")), 64)
	if err != nil || priceHelferF <= 0 {
		return &handler.ValidationError{Message: "Helfer-Preis muss größer als 0 sein"}
	}

	priceBarteamer := int64(math.Round(priceBarteamerF * 100))
	priceHelfer := int64(math.Round(priceHelferF * 100))

	ctx := r.Context()
	db := h.Store.DB()

	if err := store.UpdateItem(ctx, db, id, name, priceBarteamer, priceHelfer); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return &handler.NotFoundError{Message: "Artikel nicht gefunden"}
		}
		return fmt.Errorf("update item: %w", err)
	}

	if err := h.renderAdminCategoryList(w, r); err != nil {
		return err
	}

	h.Renderer.AppendOOB(w, "toast", map[string]string{
		"Type":    "success",
		"Message": fmt.Sprintf("Artikel '%s' aktualisiert", name),
	})

	return nil
}

// SoftDeleteItem handles POST /admin/items/{id}/delete to soft-delete a drink item.
func (h *Handler) SoftDeleteItem(w http.ResponseWriter, r *http.Request) error {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return &handler.NotFoundError{Message: "Artikel nicht gefunden"}
	}

	ctx := r.Context()
	db := h.Store.DB()

	// Fetch item name for toast before deleting.
	item, err := store.GetItem(ctx, db, id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return &handler.NotFoundError{Message: "Artikel nicht gefunden"}
		}
		return fmt.Errorf("soft delete item: get: %w", err)
	}
	if item.DeletedAt != nil {
		return &handler.NotFoundError{Message: "Artikel nicht gefunden"}
	}

	if err := store.SoftDeleteItem(ctx, db, id); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return &handler.NotFoundError{Message: "Artikel nicht gefunden"}
		}
		return fmt.Errorf("soft delete item: %w", err)
	}

	if err := h.renderAdminCategoryList(w, r); err != nil {
		return err
	}

	h.Renderer.AppendOOB(w, "toast", map[string]string{
		"Type":    "success",
		"Message": fmt.Sprintf("Artikel '%s' gelöscht", item.Name),
	})

	return nil
}
