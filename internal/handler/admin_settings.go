package handler

import (
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

// AdminSettingsPageData is the view model for the admin settings page.
type AdminSettingsPageData struct {
	User       *auth.RequestUser
	Settings   *model.Settings
	CSRFToken  string
	ActivePage string
}

// AdminSettingsPage renders the admin settings page.
func (h *Handler) AdminSettingsPage(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	user := auth.UserFromContext(ctx)
	db := h.Store.DB()

	settings, err := store.GetSettings(ctx, db)
	if err != nil {
		return fmt.Errorf("admin settings: get settings: %w", err)
	}

	data := AdminSettingsPageData{
		User:       user,
		Settings:   settings,
		CSRFToken:  middleware.CSRFTokenFromContext(ctx),
		ActivePage: "admin-settings",
	}

	h.Renderer.Page(w, r, "admin_settings", data)
	return nil
}

// parseEuroToCents parses a Euro string (e.g. "12.50") and returns cents as int64.
func parseEuroToCents(s string) (int64, error) {
	f, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return 0, err
	}
	return int64(math.Round(f * 100)), nil
}

// SaveSettings handles POST /admin/settings to update all settings.
func (h *Handler) SaveSettings(w http.ResponseWriter, r *http.Request) error {
	// Parse Euro fields to cents.
	warningLimit, err := parseEuroToCents(r.FormValue("warning_limit"))
	if err != nil {
		return &ValidationError{Message: "Ungültiger Wert für Warnlimit"}
	}
	hardSpendingLimit, err := parseEuroToCents(r.FormValue("hard_spending_limit"))
	if err != nil {
		return &ValidationError{Message: "Ungültiger Wert für Ausgabelimit"}
	}
	customTxMin, err := parseEuroToCents(r.FormValue("custom_tx_min"))
	if err != nil {
		return &ValidationError{Message: "Ungültiger Wert für Eigenbuchung Min"}
	}
	customTxMax, err := parseEuroToCents(r.FormValue("custom_tx_max"))
	if err != nil {
		return &ValidationError{Message: "Ungültiger Wert für Eigenbuchung Max"}
	}

	// Parse integer fields.
	maxItemQuantity, err := strconv.Atoi(r.FormValue("max_item_quantity"))
	if err != nil || maxItemQuantity < 1 {
		return &ValidationError{Message: "Max Anzahl muss mindestens 1 sein"}
	}
	cancellationMinutes, err := strconv.Atoi(r.FormValue("cancellation_minutes"))
	if err != nil || cancellationMinutes < 0 {
		return &ValidationError{Message: "Stornofrist darf nicht negativ sein"}
	}
	paginationSize, err := strconv.Atoi(r.FormValue("pagination_size"))
	if err != nil || paginationSize < 1 {
		return &ValidationError{Message: "Seitengröße muss mindestens 1 sein"}
	}

	// Parse SMTP fields.
	smtpPort, err := strconv.Atoi(r.FormValue("smtp_port"))
	if err != nil || smtpPort < 1 || smtpPort > 65535 {
		return &ValidationError{Message: "SMTP Port muss zwischen 1 und 65535 liegen"}
	}

	// Checkbox: present means true, absent means false.
	hardLimitEnabled := r.FormValue("hard_limit_enabled") == "true"

	// Store limits as negative values (convention: limits are negative cents).
	s := &model.Settings{
		WarningLimit:        warningLimit,
		HardSpendingLimit:   hardSpendingLimit,
		HardLimitEnabled:    hardLimitEnabled,
		CustomTxMin:         customTxMin,
		CustomTxMax:         customTxMax,
		MaxItemQuantity:     maxItemQuantity,
		CancellationMinutes: cancellationMinutes,
		PaginationSize:      paginationSize,
		SMTPHost:            strings.TrimSpace(r.FormValue("smtp_host")),
		SMTPPort:            smtpPort,
		SMTPUser:            strings.TrimSpace(r.FormValue("smtp_user")),
		SMTPPassword:        r.FormValue("smtp_password"),
		SMTPFrom:            strings.TrimSpace(r.FormValue("smtp_from")),
		EmailTemplate:       r.FormValue("email_template"),
	}

	ctx := r.Context()
	db := h.Store.DB()

	if err := store.UpdateSettings(ctx, db, s); err != nil {
		return fmt.Errorf("save settings: %w", err)
	}

	h.Renderer.Fragment(w, r, "toast", map[string]string{
		"Type":    "success",
		"Message": "Einstellungen gespeichert",
	})
	return nil
}
