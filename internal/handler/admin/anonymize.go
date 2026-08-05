package admin

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/schliz/deckel/internal/auth"
	"github.com/schliz/deckel/internal/handler"
	"github.com/schliz/deckel/internal/middleware"
	"github.com/schliz/deckel/internal/service/anonymize"
	"github.com/schliz/deckel/internal/store"
)

// AnonymizeModal renders a confirmation modal previewing how many
// transactions (and users) are due for anonymization under the currently
// configured retention period.
func (h *Handler) AnonymizeModal(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	db := h.Store.DB()

	settings, err := store.GetSettings(ctx, db)
	if err != nil {
		return fmt.Errorf("anonymize modal: get settings: %w", err)
	}

	cutoff, txCount, userCount, err := anonymize.Preview(ctx, db, settings.AnonymizeAfterMonths)
	if err != nil {
		if errors.Is(err, anonymize.ErrDisabled) {
			return &handler.ValidationError{Message: "Anonymisierung ist deaktiviert"}
		}
		return fmt.Errorf("anonymize modal: preview: %w", err)
	}

	data := map[string]any{
		"Title": "Anonymisierung",
		"Message": fmt.Sprintf(
			"%d Transaktionen von %d Nutzern (älter als %d Monate, Stichtag %s) anonymisieren? Buchungsdaten bleiben erhalten, der Personenbezug wird entfernt.",
			txCount, userCount, settings.AnonymizeAfterMonths, cutoff.Format("02.01.2006"),
		),
		"PostURL":   "/admin/settings/anonymize",
		"CSRFToken": middleware.CSRFTokenFromContext(ctx),
		"Count":     txCount,
	}

	h.Renderer.Fragment(w, r, "confirm-anonymize-modal", data)
	return nil
}

// RunAnonymization executes an anonymization run using the currently
// configured retention period, credits affected users' balance_carryover,
// and logs the run for the GoBD audit trail.
func (h *Handler) RunAnonymization(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	db := h.Store.DB()
	reqUser := auth.UserFromContext(ctx)

	settings, err := store.GetSettings(ctx, db)
	if err != nil {
		return fmt.Errorf("run anonymization: get settings: %w", err)
	}

	result, err := anonymize.Run(ctx, h.Store, settings.AnonymizeAfterMonths, reqUser.Email)
	if err != nil {
		if errors.Is(err, anonymize.ErrDisabled) {
			return &handler.ValidationError{Message: "Anonymisierung ist deaktiviert"}
		}
		return fmt.Errorf("run anonymization: %w", err)
	}

	msg := fmt.Sprintf("%d Transaktionen von %d Nutzern anonymisiert", result.TransactionCount, result.UserCount)
	if result.TransactionCount == 0 {
		msg = "Keine Transaktionen zur Anonymisierung fällig"
	}

	h.Renderer.Fragment(w, r, "toast", map[string]string{
		"Type":    "success",
		"Message": msg,
	})
	return nil
}
