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

// CreateCategory handles POST /admin/categories to create a new drink category.
func (h *Handler) CreateCategory(w http.ResponseWriter, r *http.Request) error {
	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		return &ValidationError{Message: "Kategoriename darf nicht leer sein"}
	}

	ctx := r.Context()
	db := h.Store.DB()

	_, err := store.CreateCategory(ctx, db, name)
	if err != nil {
		return fmt.Errorf("create category: %w", err)
	}

	// Render category list as main content (targeted by hx-target).
	if err := h.renderAdminCategoryList(w, r); err != nil {
		return err
	}

	// Append toast as OOB.
	h.Renderer.AppendOOB(w, "toast", map[string]string{
		"Type":    "success",
		"Message": fmt.Sprintf("Kategorie '%s' erstellt", name),
	})

	return nil
}

// CreateItem handles POST /admin/categories/{cat_id}/items to add a new drink item.
func (h *Handler) CreateItem(w http.ResponseWriter, r *http.Request) error {
	catIDStr := r.PathValue("id")
	catID, err := strconv.ParseInt(catIDStr, 10, 64)
	if err != nil {
		return &NotFoundError{Message: "Kategorie nicht gefunden"}
	}

	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		return &ValidationError{Message: "Artikelname darf nicht leer sein"}
	}

	priceBarteamerF, err := strconv.ParseFloat(r.FormValue("price_barteamer"), 64)
	if err != nil || priceBarteamerF <= 0 {
		return &ValidationError{Message: "Barteamer-Preis muss größer als 0 sein"}
	}
	priceHelferF, err := strconv.ParseFloat(r.FormValue("price_helfer"), 64)
	if err != nil || priceHelferF <= 0 {
		return &ValidationError{Message: "Helfer-Preis muss größer als 0 sein"}
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

// ReorderCategory handles POST /admin/categories/{id}/reorder to change category order.
func (h *Handler) ReorderCategory(w http.ResponseWriter, r *http.Request) error {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return &NotFoundError{Message: "Kategorie nicht gefunden"}
	}

	direction := r.URL.Query().Get("direction")
	if direction == "" {
		direction = r.FormValue("direction")
	}

	ctx := r.Context()
	db := h.Store.DB()

	if err := store.ReorderCategory(ctx, db, id, direction); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return &NotFoundError{Message: "Kategorie nicht gefunden"}
		}
		return fmt.Errorf("reorder category: %w", err)
	}

	return h.renderAdminCategoryList(w, r)
}

// DeleteCategory handles DELETE /admin/categories/{id} to remove an empty category.
func (h *Handler) DeleteCategory(w http.ResponseWriter, r *http.Request) error {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return &NotFoundError{Message: "Kategorie nicht gefunden"}
	}

	ctx := r.Context()
	db := h.Store.DB()

	// Fetch category to get its name for the toast message.
	cat, err := store.GetCategory(ctx, db, id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return &NotFoundError{Message: "Kategorie nicht gefunden"}
		}
		return fmt.Errorf("delete category: get: %w", err)
	}

	// Validate category has zero active items.
	items, err := store.ListItemsByCategory(ctx, db, id)
	if err != nil {
		return fmt.Errorf("delete category: list items: %w", err)
	}
	if len(items) > 0 {
		return &ValidationError{Message: "Kategorie kann nicht gelöscht werden, da sie noch Artikel enthält"}
	}

	if err := store.DeleteCategory(ctx, db, id); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return &NotFoundError{Message: "Kategorie nicht gefunden"}
		}
		return fmt.Errorf("delete category: %w", err)
	}

	// Render updated category list.
	if err := h.renderAdminCategoryList(w, r); err != nil {
		return err
	}

	// Append success toast as OOB.
	h.Renderer.AppendOOB(w, "toast", map[string]string{
		"Type":    "success",
		"Message": fmt.Sprintf("Kategorie '%s' gelöscht", cat.Name),
	})

	return nil
}

// renderAdminCategoryList re-fetches all categories with items and renders
// the admin-category-list partial as the main response content.
func (h *Handler) renderAdminCategoryList(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	user := auth.UserFromContext(ctx)
	db := h.Store.DB()

	settings, err := store.GetSettings(ctx, db)
	if err != nil {
		return fmt.Errorf("admin category list: get settings: %w", err)
	}

	cats, err := store.ListCategories(ctx, db)
	if err != nil {
		return fmt.Errorf("admin category list: list categories: %w", err)
	}

	var categories []CategoryWithItems
	for _, cat := range cats {
		items, err := store.ListItemsByCategory(ctx, db, cat.ID)
		if err != nil {
			return fmt.Errorf("admin category list: list items for category %d: %w", cat.ID, err)
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

	h.Renderer.Fragment(w, r, "admin-category-list", data)
	return nil
}
