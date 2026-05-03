package handler

import (
	"crypto/rand"
	"encoding/csv"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/schliz/deckel/internal/auth"
	"github.com/schliz/deckel/internal/middleware"
	"github.com/schliz/deckel/internal/model"
	"github.com/schliz/deckel/internal/store"
)

// MenuBatchSession holds the parsed diff between upload and apply.
type MenuBatchSession struct {
	Token     string
	CreatedAt time.Time
	Changes   []ItemChange
	Warnings  []string
}

// ItemChange describes a single item's old vs new values.
type ItemChange struct {
	ItemID                int64
	CategoryName          string
	OldName               string
	NewName               string
	NameChanged           bool
	OldPriceBarteamer     int64
	NewPriceBarteamer     int64
	PriceBarteamerChanged bool
	OldPriceHelfer        int64
	NewPriceHelfer        int64
	PriceHelferChanged    bool
}

// MenuBatchPageData is the view model for the menu batch edit page.
type MenuBatchPageData struct {
	User              *auth.RequestUser
	Categories        []model.Category
	Settings          *model.Settings
	CSRFToken         string
	ActivePage        string
	LowBalanceWarning bool
}

// MenuBatchDiffData is the view model for the diff preview fragment.
type MenuBatchDiffData struct {
	SessionToken string
	Changes      []ItemChange
	Warnings     []string
	CSRFToken    string
}

// MenuBatchResultData is the view model for the apply result fragment.
type MenuBatchResultData struct {
	Success      bool
	UpdatedCount int
	ErrorMessage string
}

// MenuBatchPage renders the menu batch edit wizard page.
func (h *Base) MenuBatchPage(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	user := auth.UserFromContext(ctx)
	db := h.Store.DB()

	settings, err := store.GetSettings(ctx, db)
	if err != nil {
		return fmt.Errorf("menu batch: get settings: %w", err)
	}

	cats, err := store.ListCategories(ctx, db)
	if err != nil {
		return fmt.Errorf("menu batch: list categories: %w", err)
	}

	data := MenuBatchPageData{
		User:              user,
		Categories:        cats,
		Settings:          settings,
		CSRFToken:         middleware.CSRFTokenFromContext(ctx),
		ActivePage:        "admin-menu",
		LowBalanceWarning: IsLowBalance(user, settings),
	}

	h.Renderer.Page(w, r, "admin_menu_batch", data)
	return nil
}

// MenuBatchExport exports menu items as a semicolon-separated CSV file.
func (h *Base) MenuBatchExport(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	db := h.Store.DB()

	categoryFilter := r.URL.Query().Get("category")

	cats, err := store.ListCategories(ctx, db)
	if err != nil {
		return fmt.Errorf("menu batch export: list categories: %w", err)
	}

	// Build a map of category ID → name.
	catMap := make(map[int64]string, len(cats))
	for _, c := range cats {
		catMap[c.ID] = c.Name
	}

	// Collect items based on filter.
	var items []model.Item
	if categoryFilter == "" || categoryFilter == "all" {
		for _, c := range cats {
			catItems, err := store.ListItemsByCategory(ctx, db, c.ID)
			if err != nil {
				return fmt.Errorf("menu batch export: list items for category %d: %w", c.ID, err)
			}
			items = append(items, catItems...)
		}
	} else {
		catID, err := strconv.ParseInt(categoryFilter, 10, 64)
		if err != nil {
			return &ValidationError{Message: "Ungültige Kategorie-ID"}
		}
		items, err = store.ListItemsByCategory(ctx, db, catID)
		if err != nil {
			return fmt.Errorf("menu batch export: list items: %w", err)
		}
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename=getraenkekarte.csv`)

	// Write UTF-8 BOM for Excel compatibility.
	w.Write([]byte{0xEF, 0xBB, 0xBF})

	cw := csv.NewWriter(w)
	cw.Comma = ';'

	// Header row.
	cw.Write([]string{"id", "kategorie", "name", "preis_barteamer", "preis_helfer"})

	for _, item := range items {
		catName := catMap[item.CategoryID]
		cw.Write([]string{
			strconv.FormatInt(item.ID, 10),
			catName,
			item.Name,
			formatEuroCentsExport(item.PriceBarteamer),
			formatEuroCentsExport(item.PriceHelfer),
		})
	}

	cw.Flush()
	return cw.Error()
}

// MenuBatchUpload parses an uploaded CSV, computes a diff, and shows a preview.
func (h *Base) MenuBatchUpload(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	db := h.Store.DB()

	// Clean up expired sessions.
	h.cleanupMenuBatchSessions()

	if err := r.ParseMultipartForm(1 << 20); err != nil {
		return &ValidationError{Message: "Datei zu groß (max. 1 MB)"}
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		return &ValidationError{Message: "Keine Datei ausgewählt"}
	}
	defer file.Close()

	cr := csv.NewReader(file)
	cr.Comma = ';'
	cr.LazyQuotes = true

	records, err := cr.ReadAll()
	if err != nil {
		return &ValidationError{Message: "CSV konnte nicht gelesen werden: " + err.Error()}
	}

	if len(records) < 2 {
		return &ValidationError{Message: "CSV-Datei ist leer oder enthält nur die Kopfzeile"}
	}

	// Validate header.
	header := records[0]
	if len(header) < 5 {
		return &ValidationError{Message: "CSV-Kopfzeile muss mindestens 5 Spalten haben (id;kategorie;name;preis_barteamer;preis_helfer)"}
	}

	// Build category name map for display.
	cats, err := store.ListCategories(ctx, db)
	if err != nil {
		return fmt.Errorf("menu batch upload: list categories: %w", err)
	}
	catMap := make(map[int64]string, len(cats))
	for _, c := range cats {
		catMap[c.ID] = c.Name
	}

	var changes []ItemChange
	var warnings []string

	for i, row := range records[1:] {
		lineNum := i + 2 // 1-based, skip header

		if len(row) < 5 {
			warnings = append(warnings, fmt.Sprintf("Zeile %d: Zu wenige Spalten, übersprungen", lineNum))
			continue
		}

		// Strip BOM from first field if present.
		idStr := strings.TrimPrefix(strings.TrimSpace(row[0]), "\xEF\xBB\xBF")
		itemID, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("Zeile %d: Ungültige ID '%s'", lineNum, idStr))
			continue
		}

		name := strings.TrimSpace(row[2])
		if name == "" {
			warnings = append(warnings, fmt.Sprintf("Zeile %d: Name darf nicht leer sein", lineNum))
			continue
		}

		priceBarteamer, err := parseMenuBatchPrice(row[3])
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("Zeile %d: Ungültiger Preis '%s'", lineNum, strings.TrimSpace(row[3])))
			continue
		}
		if priceBarteamer <= 0 {
			warnings = append(warnings, fmt.Sprintf("Zeile %d: Preis muss größer als 0 sein", lineNum))
			continue
		}

		priceHelfer, err := parseMenuBatchPrice(row[4])
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("Zeile %d: Ungültiger Preis '%s'", lineNum, strings.TrimSpace(row[4])))
			continue
		}
		if priceHelfer <= 0 {
			warnings = append(warnings, fmt.Sprintf("Zeile %d: Preis muss größer als 0 sein", lineNum))
			continue
		}

		// Look up existing item.
		item, err := store.GetItem(ctx, db, itemID)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				warnings = append(warnings, fmt.Sprintf("Zeile %d: Unbekannte ID %d", lineNum, itemID))
				continue
			}
			return fmt.Errorf("menu batch upload: get item %d: %w", itemID, err)
		}

		if item.DeletedAt != nil {
			warnings = append(warnings, fmt.Sprintf("Zeile %d: Artikel wurde gelöscht (ID %d)", lineNum, itemID))
			continue
		}

		// Compute diff.
		change := ItemChange{
			ItemID:                itemID,
			CategoryName:          catMap[item.CategoryID],
			OldName:               item.Name,
			NewName:               name,
			NameChanged:           item.Name != name,
			OldPriceBarteamer:     item.PriceBarteamer,
			NewPriceBarteamer:     priceBarteamer,
			PriceBarteamerChanged: item.PriceBarteamer != priceBarteamer,
			OldPriceHelfer:        item.PriceHelfer,
			NewPriceHelfer:        priceHelfer,
			PriceHelferChanged:    item.PriceHelfer != priceHelfer,
		}

		if change.NameChanged || change.PriceBarteamerChanged || change.PriceHelferChanged {
			changes = append(changes, change)
		}
	}

	// Generate session token.
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return fmt.Errorf("menu batch upload: generate token: %w", err)
	}
	token := fmt.Sprintf("%x", tokenBytes)

	session := &MenuBatchSession{
		Token:     token,
		CreatedAt: time.Now(),
		Changes:   changes,
		Warnings:  warnings,
	}
	h.MenuBatchSessions.Store(token, session)

	data := MenuBatchDiffData{
		SessionToken: token,
		Changes:      changes,
		Warnings:     warnings,
		CSRFToken:    middleware.CSRFTokenFromContext(ctx),
	}

	h.Renderer.Fragment(w, r, "menu-batch-diff-preview", data)
	return nil
}

// MenuBatchApply applies the changes from a menu batch upload session.
func (h *Base) MenuBatchApply(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	token := r.FormValue("session_token")
	val, ok := h.MenuBatchSessions.LoadAndDelete(token)
	if !ok {
		return &ValidationError{Message: "Sitzung abgelaufen. Bitte erneut hochladen."}
	}

	session := val.(*MenuBatchSession)

	// Check session age.
	if time.Since(session.CreatedAt) > 30*time.Minute {
		return &ValidationError{Message: "Sitzung abgelaufen. Bitte erneut hochladen."}
	}

	if len(session.Changes) == 0 {
		data := MenuBatchResultData{
			Success:      true,
			UpdatedCount: 0,
		}
		h.Renderer.Fragment(w, r, "menu-batch-apply-result", data)
		return nil
	}

	var updatedCount int
	err := h.Store.WithTx(ctx, func(tx pgx.Tx) error {
		for _, change := range session.Changes {
			// Verify item still exists and is not deleted.
			item, err := store.GetItem(ctx, tx, change.ItemID)
			if err != nil {
				return fmt.Errorf("Artikel ID %d nicht mehr gefunden", change.ItemID)
			}
			if item.DeletedAt != nil {
				return fmt.Errorf("Artikel '%s' (ID %d) wurde zwischenzeitlich gelöscht", item.Name, change.ItemID)
			}

			if err := store.UpdateItem(ctx, tx, change.ItemID, change.NewName, change.NewPriceBarteamer, change.NewPriceHelfer); err != nil {
				return fmt.Errorf("Fehler beim Aktualisieren von '%s': %w", change.NewName, err)
			}
			updatedCount++
		}
		return nil
	})

	if err != nil {
		data := MenuBatchResultData{
			Success:      false,
			ErrorMessage: err.Error(),
		}
		h.Renderer.Fragment(w, r, "menu-batch-apply-result", data)
		return nil
	}

	data := MenuBatchResultData{
		Success:      true,
		UpdatedCount: updatedCount,
	}
	h.Renderer.Fragment(w, r, "menu-batch-apply-result", data)
	return nil
}

// parseMenuBatchPrice parses a German-locale price string like "1,50" into cents (150).
func parseMenuBatchPrice(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty price")
	}
	s = NormalizeDecimal(s)
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}
	return int64(math.Round(f * 100)), nil
}

// cleanupMenuBatchSessions removes sessions older than 30 minutes.
func (h *Base) cleanupMenuBatchSessions() {
	h.MenuBatchSessions.Range(func(key, value any) bool {
		session := value.(*MenuBatchSession)
		if time.Since(session.CreatedAt) > 30*time.Minute {
			h.MenuBatchSessions.Delete(key)
		}
		return true
	})
}
