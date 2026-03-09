package store

import (
	"context"
	"fmt"

	"github.com/schliz/deckel/internal/model"
)

// GetSettings reads the singleton settings row (id=1).
func GetSettings(ctx context.Context, db DBTX) (*model.Settings, error) {
	var s model.Settings
	err := db.QueryRow(ctx, `
		SELECT id, warning_limit, hard_spending_limit, hard_limit_enabled,
		       custom_tx_min, custom_tx_max, max_item_quantity,
		       cancellation_minutes, pagination_size,
		       smtp_host, smtp_port, smtp_user, smtp_password, smtp_from,
		       email_template, updated_at
		FROM settings WHERE id = 1`).Scan(
		&s.ID, &s.WarningLimit, &s.HardSpendingLimit, &s.HardLimitEnabled,
		&s.CustomTxMin, &s.CustomTxMax, &s.MaxItemQuantity,
		&s.CancellationMinutes, &s.PaginationSize,
		&s.SMTPHost, &s.SMTPPort, &s.SMTPUser, &s.SMTPPassword, &s.SMTPFrom,
		&s.EmailTemplate, &s.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get settings: %w", err)
	}
	return &s, nil
}

// UpdateSettings updates all fields of the singleton settings row and sets updated_at.
func UpdateSettings(ctx context.Context, db DBTX, s *model.Settings) error {
	_, err := db.Exec(ctx, `
		UPDATE settings SET
			warning_limit = $1,
			hard_spending_limit = $2,
			hard_limit_enabled = $3,
			custom_tx_min = $4,
			custom_tx_max = $5,
			max_item_quantity = $6,
			cancellation_minutes = $7,
			pagination_size = $8,
			smtp_host = $9,
			smtp_port = $10,
			smtp_user = $11,
			smtp_password = $12,
			smtp_from = $13,
			email_template = $14,
			updated_at = NOW()
		WHERE id = 1`,
		s.WarningLimit, s.HardSpendingLimit, s.HardLimitEnabled,
		s.CustomTxMin, s.CustomTxMax, s.MaxItemQuantity,
		s.CancellationMinutes, s.PaginationSize,
		s.SMTPHost, s.SMTPPort, s.SMTPUser, s.SMTPPassword, s.SMTPFrom,
		s.EmailTemplate,
	)
	if err != nil {
		return fmt.Errorf("update settings: %w", err)
	}
	return nil
}
