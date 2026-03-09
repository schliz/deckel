package handler

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/schliz/deckel/internal/auth"
	"github.com/schliz/deckel/internal/middleware"
	"github.com/schliz/deckel/internal/store"
)

// ProfilePageData is the view model for the profile page.
type ProfilePageData struct {
	User              *auth.RequestUser
	CSRFToken         string
	ActivePage        string
	LowBalanceWarning bool
}

// ProfilePage renders the user profile page.
func (h *Handler) ProfilePage(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	user := auth.UserFromContext(ctx)

	settings, err := store.GetSettings(ctx, h.Store.DB())
	if err != nil {
		return fmt.Errorf("profile: get settings: %w", err)
	}

	data := ProfilePageData{
		User:              user,
		CSRFToken:         middleware.CSRFTokenFromContext(ctx),
		ActivePage:        "profile",
		LowBalanceWarning: isLowBalance(user, settings),
	}

	h.Renderer.Page(w, r, "profile", data)
	return nil
}

// ExportData generates a GDPR data export as a ZIP file containing profile.json and transactions.csv.
func (h *Handler) ExportData(w http.ResponseWriter, r *http.Request) error {
	user := auth.UserFromContext(r.Context())
	ctx := r.Context()

	// Fetch all transactions for this user.
	txns, err := store.ListAllTransactionsByUser(ctx, h.Store.DB(), user.ID)
	if err != nil {
		return fmt.Errorf("export data: list transactions: %w", err)
	}

	// Build profile JSON.
	profile := map[string]any{
		"email":       user.Email,
		"full_name":   user.FullName,
		"given_name":  user.GivenName,
		"family_name": user.FamilyName,
		"is_barteamer": user.IsBarteamer,
		"is_admin":    user.IsAdmin,
		"is_active":   user.IsActive,
		"balance":     user.Balance,
		"created_at":  user.CreatedAt,
	}
	profileJSON, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return fmt.Errorf("export data: marshal profile: %w", err)
	}

	// Build transactions CSV with semicolon separator (German locale).
	csv := "Datum;Beschreibung;Menge;Betrag;Typ;Status\n"
	for _, t := range txns {
		datum := t.CreatedAt.Format("02.01.2006 15:04")

		beschreibung := ""
		if t.Description != nil {
			beschreibung = *t.Description
		} else if t.ItemTitle != nil {
			beschreibung = *t.ItemTitle
		}

		menge := ""
		if t.Quantity != nil {
			menge = fmt.Sprintf("%d", *t.Quantity)
		}

		betrag := formatEuroCentsExport(t.Amount)

		status := "Aktiv"
		if t.CancelledAt != nil {
			status = "Storniert"
		}

		csv += fmt.Sprintf("%s;%s;%s;%s;%s;%s\n",
			datum, escapeSemicolon(beschreibung), menge, betrag, t.Type, status)
	}

	// Write ZIP response.
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", `attachment; filename=deckel-export.zip`)

	zw := zip.NewWriter(w)

	// Add profile.json.
	pw, err := zw.Create("profile.json")
	if err != nil {
		return fmt.Errorf("export data: create profile.json: %w", err)
	}
	if _, err := pw.Write(profileJSON); err != nil {
		return fmt.Errorf("export data: write profile.json: %w", err)
	}

	// Add transactions.csv.
	tw, err := zw.Create("transactions.csv")
	if err != nil {
		return fmt.Errorf("export data: create transactions.csv: %w", err)
	}
	if _, err := tw.Write([]byte(csv)); err != nil {
		return fmt.Errorf("export data: write transactions.csv: %w", err)
	}

	if err := zw.Close(); err != nil {
		return fmt.Errorf("export data: close zip: %w", err)
	}

	return nil
}

// formatEuroCentsExport formats cents as "X,YY" for CSV export.
func formatEuroCentsExport(cents int64) string {
	negative := cents < 0
	if negative {
		cents = -cents
	}
	prefix := ""
	if negative {
		prefix = "-"
	}
	return fmt.Sprintf("%s%d,%02d", prefix, cents/100, cents%100)
}

// escapeSemicolon wraps a string in quotes if it contains semicolons.
func escapeSemicolon(s string) string {
	for _, c := range s {
		if c == ';' || c == '"' || c == '\n' {
			// Escape quotes by doubling them, wrap in quotes
			escaped := ""
			for _, ch := range s {
				if ch == '"' {
					escaped += `""`
				} else {
					escaped += string(ch)
				}
			}
			return `"` + escaped + `"`
		}
	}
	return s
}
