package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/k4-bar/deckel/internal/auth"
	"github.com/k4-bar/deckel/internal/middleware"
	"github.com/k4-bar/deckel/internal/model"
	"github.com/k4-bar/deckel/internal/store"
)

// OrderModalData is the view model for the order confirmation modal.
type OrderModalData struct {
	Item        *model.Item
	User        *auth.RequestUser
	MaxQuantity int
	CSRFToken   string
}

// OrderModal renders the order confirmation modal for a given item.
func (h *Handler) OrderModal(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	user := auth.UserFromContext(ctx)
	db := h.Store.DB()

	// Extract item ID from URL path.
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return &NotFoundError{Message: "Artikel nicht gefunden"}
	}

	// Fetch item from DB.
	item, err := store.GetItem(ctx, db, id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return &NotFoundError{Message: "Artikel nicht gefunden"}
		}
		return fmt.Errorf("order modal: get item: %w", err)
	}

	// Return 404 if soft-deleted.
	if item.DeletedAt != nil {
		return &NotFoundError{Message: "Artikel nicht gefunden"}
	}

	// Fetch settings for max_item_quantity.
	settings, err := store.GetSettings(ctx, db)
	if err != nil {
		return fmt.Errorf("order modal: get settings: %w", err)
	}

	data := OrderModalData{
		Item:        item,
		User:        user,
		MaxQuantity: settings.MaxItemQuantity,
		CSRFToken:   middleware.CSRFTokenFromContext(ctx),
	}

	h.Renderer.Fragment(w, r, "order-modal", data)
	return nil
}

// PlaceOrder processes an order submission: validates input, checks spending limits,
// creates a purchase transaction, and responds with toast + OOB header-stats update.
func (h *Handler) PlaceOrder(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	user := auth.UserFromContext(ctx)

	// Parse form data.
	itemIDStr := r.FormValue("item_id")
	qtyStr := r.FormValue("quantity")

	itemID, err := strconv.ParseInt(itemIDStr, 10, 64)
	if err != nil {
		return &ValidationError{Message: "Ungültige Artikel-ID"}
	}

	quantity, err := strconv.Atoi(qtyStr)
	if err != nil || quantity < 1 {
		return &ValidationError{Message: "Ungültige Menge"}
	}

	db := h.Store.DB()

	// Fetch item.
	item, err := store.GetItem(ctx, db, itemID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return &NotFoundError{Message: "Artikel nicht gefunden"}
		}
		return fmt.Errorf("place order: get item: %w", err)
	}
	if item.DeletedAt != nil {
		return &NotFoundError{Message: "Artikel nicht gefunden"}
	}

	// Fetch settings for max quantity and spending limits.
	settings, err := store.GetSettings(ctx, db)
	if err != nil {
		return fmt.Errorf("place order: get settings: %w", err)
	}

	// Validate quantity against max.
	if quantity > settings.MaxItemQuantity {
		return &ValidationError{Message: fmt.Sprintf("Maximal %d Stück erlaubt", settings.MaxItemQuantity)}
	}

	// Determine price tier based on user role.
	var unitPrice int64
	if user.IsBarteamer {
		unitPrice = item.PriceBarteamer
	} else {
		unitPrice = item.PriceHelfer
	}

	// Amount is negative (purchase debits the user's tab).
	amount := -(unitPrice * int64(quantity))

	var warning bool

	// Execute within a DB transaction with FOR UPDATE locking.
	err = h.Store.WithTx(ctx, func(tx pgx.Tx) error {
		// Get current balance with row lock.
		balance, err := store.GetUserBalanceForUpdate(ctx, tx, user.ID)
		if err != nil {
			return fmt.Errorf("get balance for update: %w", err)
		}

		// Check hard spending limit (if enabled and user not exempt).
		if settings.HardLimitEnabled && !user.SpendingLimitDisabled {
			hardLimit := -settings.HardSpendingLimit
			if balance <= hardLimit {
				return &ValidationError{Message: "Bestellung nicht möglich: Ausgabenlimit erreicht. Bitte erst einzahlen."}
			}
			projectedBalance := balance + amount
			if projectedBalance <= hardLimit {
				warning = true
			}
		}

		// Create purchase transaction.
		itemTitle := item.Name
		txn := &model.Transaction{
			UserID:    user.ID,
			Amount:    amount,
			ItemTitle: &itemTitle,
			UnitPrice: &unitPrice,
			Quantity:  &quantity,
			Type:      "purchase",
		}
		_, err = store.CreateTransaction(ctx, tx, txn)
		return err
	})
	if err != nil {
		// Check if it's a ValidationError from inside the tx.
		var valErr *ValidationError
		if errors.As(err, &valErr) {
			return valErr
		}
		return fmt.Errorf("place order: %w", err)
	}

	// Build response: success overlay + optional warning toast + OOB header-stats.
	h.Renderer.Fragment(w, r, "order-success", map[string]any{
		"Quantity":    quantity,
		"ItemName":    item.Name,
		"TotalAmount": -amount,
	})

	// Show additional warning toast when this order hits the spending limit.
	if warning {
		h.Renderer.AppendOOB(w, "toast", map[string]string{
			"Type":    "warning",
			"Message": "Achtung: Du hast dein Limit erreicht. Bitte zahle bald ein!",
		})
	}

	// Render OOB header-stats update.
	newBalance, _ := store.GetUserBalance(ctx, db, user.ID)
	totalBalance, _ := store.GetAllBalancesSum(ctx, db)
	negativeSum, _ := store.GetNegativeBalancesSum(ctx, db)
	rank, total, _ := store.GetUserRank(ctx, db, user.ID)

	h.Renderer.AppendOOB(w, "header-stats", map[string]any{
		"UserBalance":         newBalance,
		"TotalBalance":        totalBalance,
		"NegativeBalancesSum": negativeSum,
		"UserRank":            rank,
		"TotalUsers":          total,
		"Settings":            settings,
		"User":                user,
		"OOB":                 true,
	})

	return nil
}
