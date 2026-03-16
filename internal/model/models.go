package model

import "time"

type User struct {
	ID                   int64
	Email                string
	FullName             string
	GivenName            string
	FamilyName           string
	IsBarteamer          bool
	IsAdmin              bool
	IsKiosk              bool
	IsActive             bool
	SpendingLimitDisabled bool
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type UserWithBalance struct {
	User
	Balance int64
}

type Category struct {
	ID        int64
	Name      string
	SortOrder int
	CreatedAt time.Time
}

type Item struct {
	ID             int64
	CategoryID     int64
	Name           string
	PriceBarteamer int64
	PriceHelfer    int64
	SortOrder      int
	DeletedAt      *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type Transaction struct {
	ID          int64
	UserID      int64
	Amount      int64
	ItemTitle   *string
	UnitPrice   *int64
	Quantity    *int
	Description *string
	Type        string
	CancelledAt *time.Time
	CancelsID       *int64
	CreatedByUserID *int64
	CreatedAt       time.Time
}

type TransactionWithUser struct {
	Transaction
	UserName  string
	UserEmail string
}

type Settings struct {
	ID                 int
	WarningLimit       int64
	HardSpendingLimit  int64
	HardLimitEnabled   bool
	CustomTxMin        int64
	CustomTxMax        int64
	MaxItemQuantity    int
	CancellationMinutes int
	PaginationSize     int
	SMTPHost           string
	SMTPPort           int
	SMTPUser           string
	SMTPPassword       string
	SMTPFrom           string
	SMTPFromName       string
	EmailSubject       string
	EmailTemplate      string
	UpdatedAt          time.Time
}
