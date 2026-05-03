package handler

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/schliz/deckel/internal/auth"
	"github.com/schliz/deckel/internal/config"
	"github.com/schliz/deckel/internal/model"
	"github.com/schliz/deckel/internal/render"
	"github.com/schliz/deckel/internal/store"
)

// AppHandler is a handler function that returns an error for centralized error handling.
type AppHandler func(w http.ResponseWriter, r *http.Request) error

// Base holds shared dependencies for all HTTP handlers.
type Base struct {
	Store       *store.Store
	Renderer    *render.Renderer
	Config      *config.Config
	MenuBatchSessions sync.Map
}

// NotFoundError indicates a resource was not found.
type NotFoundError struct {
	Message string
}

func (e *NotFoundError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return "not found"
}

// ForbiddenError indicates the user does not have permission.
type ForbiddenError struct {
	Message string
}

func (e *ForbiddenError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return "forbidden"
}

// ValidationError indicates invalid user input.
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return "invalid input"
}

// Wrap converts an AppHandler into a standard http.HandlerFunc.
// If the handler returns an error, it is mapped to an appropriate HTTP response.
func (h *Base) Wrap(fn AppHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := fn(w, r); err != nil {
			h.handleError(w, r, err)
		}
	}
}

// handleError maps error types to HTTP status codes and renders an appropriate response.
func (h *Base) handleError(w http.ResponseWriter, r *http.Request, err error) {
	code := http.StatusInternalServerError
	msg := "Interner Serverfehler"
	title := "Interner Fehler"

	var notFound *NotFoundError
	var forbidden *ForbiddenError
	var validation *ValidationError

	switch {
	case errors.As(err, &notFound):
		code = http.StatusNotFound
		msg = notFound.Error()
		title = "Seite nicht gefunden"
	case errors.As(err, &forbidden):
		code = http.StatusForbidden
		msg = forbidden.Error()
		title = "Zugriff verweigert"
	case errors.As(err, &validation):
		code = http.StatusBadRequest
		msg = validation.Error()
		title = "Ungültige Eingabe"
	default:
		log.Printf("ERROR: %v", err)
	}

	if IsHTMX(r) {
		w.WriteHeader(code)
		h.Renderer.Fragment(w, r, "toast", map[string]string{
			"Type":    "error",
			"Message": msg,
		})
		return
	}

	w.WriteHeader(code)
	h.RenderErrorPage(w, r, code, title, msg)
}

// RenderErrorPage renders a styled error page for non-HTMX requests.
func (h *Base) RenderErrorPage(w http.ResponseWriter, r *http.Request, code int, title, message string) {
	data := map[string]any{
		"ErrorCode":    code,
		"ErrorTitle":   title,
		"ErrorMessage": message,
	}
	h.Renderer.Page(w, r, "error", data)
}

// NotFoundHandler returns a handler for unmatched routes.
func (h *Base) NotFoundHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if IsHTMX(r) {
			w.WriteHeader(http.StatusNotFound)
			h.Renderer.Fragment(w, r, "toast", map[string]string{
				"Type":    "error",
				"Message": "Seite nicht gefunden",
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
		h.RenderErrorPage(w, r, http.StatusNotFound, "Seite nicht gefunden", "Die angeforderte Seite wurde nicht gefunden.")
	}
}

// NormalizeDecimal replaces comma with period so that German-locale decimal
// inputs like "1,50" are accepted by strconv.ParseFloat.
func NormalizeDecimal(s string) string {
	return strings.ReplaceAll(strings.TrimSpace(s), ",", ".")
}

// ValidateTextLen validates that s is at most maxLen bytes. Returns a
// ValidationError with the given message if exceeded.
func ValidateTextLen(s string, maxLen int, field string) error {
	if len(s) > maxLen {
		return &ValidationError{Message: fmt.Sprintf("%s darf maximal %d Zeichen lang sein", field, maxLen)}
	}
	return nil
}

// IsHTMX checks if the request was made by HTMX.
func IsHTMX(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

// IsLowBalance returns true if the user's balance is below the warning limit.
func IsLowBalance(user *auth.RequestUser, settings *model.Settings) bool {
	if user == nil || settings == nil {
		return false
	}
	return user.Balance < settings.WarningLimit
}

// HeaderStats renders the header-stats component for initial page load.
func (h *Base) HeaderStats(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	user := auth.UserFromContext(ctx)
	db := h.Store.DB()

	settings, err := store.GetSettings(ctx, db)
	if err != nil {
		return fmt.Errorf("header stats: get settings: %w", err)
	}

	balance, _ := store.GetUserBalance(ctx, db, user.ID)
	totalBalance, _ := store.GetAllBalancesSum(ctx, db)
	negativeSum, _ := store.GetNegativeBalancesSum(ctx, db)
	rank, total, _ := store.GetUserRank(ctx, db, user.ID)

	h.Renderer.Fragment(w, r, "header-stats", map[string]any{
		"UserBalance":         balance,
		"TotalBalance":        totalBalance,
		"NegativeBalancesSum": negativeSum,
		"UserRank":            rank,
		"TotalUsers":          total,
		"Settings":            settings,
		"User":                user,
	})
	return nil
}
