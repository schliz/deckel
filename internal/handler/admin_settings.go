package handler

import (
	"bytes"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"text/template"

	"github.com/schliz/deckel/internal/auth"
	"github.com/schliz/deckel/internal/mail"
	"github.com/schliz/deckel/internal/middleware"
	"github.com/schliz/deckel/internal/model"
	"github.com/schliz/deckel/internal/store"
)

// AdminSettingsPageData is the view model for the admin settings page.
type AdminSettingsPageData struct {
	User              *auth.RequestUser
	Settings          *model.Settings
	CSRFToken         string
	ActivePage        string
	LowBalanceWarning bool
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
		User:              user,
		Settings:          settings,
		CSRFToken:         middleware.CSRFTokenFromContext(ctx),
		ActivePage:        "admin-settings",
		LowBalanceWarning: isLowBalance(user, settings),
	}

	h.Renderer.Page(w, r, "admin_settings", data)
	return nil
}

// parseEuroToCents parses a Euro string (e.g. "12.50" or "12,50") and returns cents as int64.
// Accepts both period and comma as decimal separator.
func parseEuroToCents(s string) (int64, error) {
	f, err := strconv.ParseFloat(normalizeDecimal(s), 64)
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
	if err != nil || maxItemQuantity < 1 || maxItemQuantity > 100 {
		return &ValidationError{Message: "Max Anzahl muss zwischen 1 und 100 liegen"}
	}
	cancellationMinutes, err := strconv.Atoi(r.FormValue("cancellation_minutes"))
	if err != nil || cancellationMinutes < 0 || cancellationMinutes > 10080 {
		return &ValidationError{Message: "Stornofrist muss zwischen 0 und 10080 Minuten liegen"}
	}
	paginationSize, err := strconv.Atoi(r.FormValue("pagination_size"))
	if err != nil || paginationSize < 1 || paginationSize > 500 {
		return &ValidationError{Message: "Seitengröße muss zwischen 1 und 500 liegen"}
	}

	// Parse SMTP fields.
	smtpPort, err := strconv.Atoi(r.FormValue("smtp_port"))
	if err != nil || smtpPort < 1 || smtpPort > 65535 {
		return &ValidationError{Message: "SMTP Port muss zwischen 1 und 65535 liegen"}
	}

	// Checkbox: present means true, absent means false.
	hardLimitEnabled := r.FormValue("hard_limit_enabled") == "true"

	// Parse and validate text fields.
	smtpHost := strings.TrimSpace(r.FormValue("smtp_host"))
	if err := validateTextLen(smtpHost, 255, "SMTP Host"); err != nil {
		return err
	}
	smtpUser := strings.TrimSpace(r.FormValue("smtp_user"))
	if err := validateTextLen(smtpUser, 255, "SMTP User"); err != nil {
		return err
	}
	smtpFrom := strings.TrimSpace(r.FormValue("smtp_from"))
	if err := validateTextLen(smtpFrom, 255, "SMTP From"); err != nil {
		return err
	}
	smtpFromName := strings.TrimSpace(r.FormValue("smtp_from_name"))
	if err := validateTextLen(smtpFromName, 255, "Absender-Name"); err != nil {
		return err
	}
	emailSubject := strings.TrimSpace(r.FormValue("email_subject"))
	if err := validateTextLen(emailSubject, 255, "Betreff"); err != nil {
		return err
	}
	emailTemplate := strings.TrimSpace(r.FormValue("email_template"))
	if err := validateTextLen(emailTemplate, 10000, "E-Mail-Template"); err != nil {
		return err
	}

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
		SMTPHost:            smtpHost,
		SMTPPort:            smtpPort,
		SMTPUser:            smtpUser,
		SMTPPassword:        r.FormValue("smtp_password"),
		SMTPFrom:            smtpFrom,
		SMTPFromName:        smtpFromName,
		EmailSubject:        emailSubject,
		EmailTemplate:       emailTemplate,
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

// SendRemindersModal renders a confirmation modal showing how many users would receive reminders.
func (h *Handler) SendRemindersModal(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	db := h.Store.DB()

	limitType := r.URL.Query().Get("type")
	if limitType != "warning" && limitType != "hard" {
		return &ValidationError{Message: "Ungültiger Typ"}
	}

	settings, err := store.GetSettings(ctx, db)
	if err != nil {
		return fmt.Errorf("send reminders modal: get settings: %w", err)
	}

	users, err := store.ListActiveUsersWithBalance(ctx, db)
	if err != nil {
		return fmt.Errorf("send reminders modal: list users: %w", err)
	}

	var threshold int64
	var title, description string
	if limitType == "warning" {
		threshold = settings.WarningLimit
		title = "Erinnerung: Warnlimit"
		description = fmt.Sprintf("unter dem Warnlimit (%s)", formatCentsToEuro(settings.WarningLimit))
	} else {
		threshold = -settings.HardSpendingLimit
		title = "Erinnerung: Ausgabelimit"
		description = fmt.Sprintf("unter dem Ausgabelimit (%s)", formatCentsToEuro(-settings.HardSpendingLimit))
	}

	var count int
	for _, u := range users {
		if u.Balance < threshold {
			count++
		}
	}

	data := map[string]any{
		"Title":     title,
		"Message":   fmt.Sprintf("Erinnerungsmail an %d Nutzer %s senden?", count, description),
		"PostURL":   "/admin/settings/send-reminders?type=" + limitType,
		"CSRFToken": middleware.CSRFTokenFromContext(ctx),
		"Count":     count,
	}

	h.Renderer.Fragment(w, r, "confirm-send-reminders-modal", data)
	return nil
}

// SendReminders sends balance reminder emails to active users below the specified limit.
func (h *Handler) SendReminders(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	db := h.Store.DB()

	limitType := r.URL.Query().Get("type")
	if limitType != "warning" && limitType != "hard" {
		return &ValidationError{Message: "Ungültiger Typ"}
	}

	users, err := store.ListActiveUsersWithBalance(ctx, db)
	if err != nil {
		return fmt.Errorf("send reminders: list users: %w", err)
	}

	settings, err := store.GetSettings(ctx, db)
	if err != nil {
		return fmt.Errorf("send reminders: get settings: %w", err)
	}

	var threshold int64
	if limitType == "warning" {
		threshold = settings.WarningLimit
	} else {
		threshold = -settings.HardSpendingLimit
	}

	tmpl, err := template.New("email").Parse(settings.EmailTemplate)
	if err != nil {
		return &ValidationError{Message: "E-Mail-Template ist ungültig: " + err.Error()}
	}

	mailer := &mail.Mailer{
		Host:     settings.SMTPHost,
		Port:     strconv.Itoa(settings.SMTPPort),
		Username: settings.SMTPUser,
		Password: settings.SMTPPassword,
		From:     settings.SMTPFrom,
		FromName: settings.SMTPFromName,
	}

	var successes, failures int
	for _, u := range users {
		if u.Balance >= threshold {
			continue
		}

		var buf bytes.Buffer
		data := map[string]string{
			"Name":      u.FullName,
			"FirstName": u.GivenName,
			"Balance":   formatCentsToEuro(u.Balance),
		}
		if err := tmpl.Execute(&buf, data); err != nil {
			log.Printf("send reminders: render template for %s: %v", u.Email, err)
			failures++
			continue
		}

		if err := mailer.Send(u.Email, settings.EmailSubject, buf.String()); err != nil {
			log.Printf("send reminders: send to %s: %v", u.Email, err)
			failures++
			continue
		}
		successes++
	}

	toastType := "success"
	if successes == 0 && failures > 0 {
		toastType = "error"
	} else if failures > 0 {
		toastType = "warning"
	}

	msg := fmt.Sprintf("%d Emails gesendet, %d Fehler", successes, failures)
	h.Renderer.Fragment(w, r, "toast", map[string]string{
		"Type":    toastType,
		"Message": msg,
	})
	return nil
}

// formatCentsToEuro formats a cent amount as a Euro string (e.g. -150 → "-1.50 EUR").
func formatCentsToEuro(cents int64) string {
	sign := ""
	if cents < 0 {
		sign = "-"
		cents = -cents
	}
	return fmt.Sprintf("%s%d.%02d EUR", sign, cents/100, cents%100)
}
