package kiosk

// Kiosk order confirmation and submission. Renders the per-user/per-item
// confirmation page and processes the actual purchase transaction.

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/schliz/deckel/internal/auth"
	"github.com/schliz/deckel/internal/handler"
	"github.com/schliz/deckel/internal/middleware"
	"github.com/schliz/deckel/internal/model"
	"github.com/schliz/deckel/internal/store"
)

// KioskConfirm renders the order confirmation page for a kiosk order.
func (h *Handler) KioskConfirm(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	db := h.Store.DB()

	itemID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		return &handler.NotFoundError{Message: "Artikel nicht gefunden"}
	}

	userID, err := strconv.ParseInt(r.PathValue("uid"), 10, 64)
	if err != nil {
		return &handler.NotFoundError{Message: "Benutzer nicht gefunden"}
	}

	item, err := store.GetItem(ctx, db, itemID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return &handler.NotFoundError{Message: "Artikel nicht gefunden"}
		}
		return fmt.Errorf("kiosk confirm: get item: %w", err)
	}
	if item.DeletedAt != nil {
		return &handler.NotFoundError{Message: "Artikel nicht gefunden"}
	}

	targetUser, err := store.GetUserWithBalance(ctx, db, userID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return &handler.NotFoundError{Message: "Benutzer nicht gefunden"}
		}
		return fmt.Errorf("kiosk confirm: get user: %w", err)
	}

	settings, err := store.GetSettings(ctx, db)
	if err != nil {
		return fmt.Errorf("kiosk confirm: get settings: %w", err)
	}

	var unitPrice int64
	if targetUser.IsBarteamer {
		unitPrice = item.PriceBarteamer
	} else {
		unitPrice = item.PriceHelfer
	}

	isBlocked := false
	if settings.HardLimitEnabled && !targetUser.SpendingLimitDisabled {
		isBlocked = targetUser.Balance <= -settings.HardSpendingLimit
	}

	isLowBalance := targetUser.Balance < settings.WarningLimit

	data := map[string]any{
		"Item":         item,
		"TargetUser":   targetUser,
		"UnitPrice":    unitPrice,
		"Settings":     settings,
		"CSRFToken":    middleware.CSRFTokenFromContext(ctx),
		"IsBlocked":    isBlocked,
		"IsLowBalance": isLowBalance,
	}

	h.Renderer.Page(w, r, "kiosk/confirm", data)
	return nil
}

// KioskPlaceOrder processes a kiosk order submission.
func (h *Handler) KioskPlaceOrder(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	kioskUser := auth.UserFromContext(ctx)
	db := h.Store.DB()

	itemID, err := strconv.ParseInt(r.FormValue("item_id"), 10, 64)
	if err != nil {
		return &handler.ValidationError{Message: "Ungültige Artikel-ID"}
	}

	userID, err := strconv.ParseInt(r.FormValue("user_id"), 10, 64)
	if err != nil {
		return &handler.ValidationError{Message: "Ungültige Benutzer-ID"}
	}

	quantity, err := strconv.Atoi(r.FormValue("quantity"))
	if err != nil || quantity < 1 {
		return &handler.ValidationError{Message: "Ungültige Menge"}
	}

	item, err := store.GetItem(ctx, db, itemID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return &handler.NotFoundError{Message: "Artikel nicht gefunden"}
		}
		return fmt.Errorf("kiosk place order: get item: %w", err)
	}
	if item.DeletedAt != nil {
		return &handler.NotFoundError{Message: "Artikel nicht gefunden"}
	}

	targetUser, err := store.GetUserWithBalance(ctx, db, userID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return &handler.NotFoundError{Message: "Benutzer nicht gefunden"}
		}
		return fmt.Errorf("kiosk place order: get user: %w", err)
	}

	if !targetUser.IsActive {
		return &handler.ValidationError{Message: "Benutzer ist deaktiviert"}
	}

	settings, err := store.GetSettings(ctx, db)
	if err != nil {
		return fmt.Errorf("kiosk place order: get settings: %w", err)
	}

	if quantity > settings.MaxItemQuantity {
		return &handler.ValidationError{Message: fmt.Sprintf("Maximal %d Stück erlaubt", settings.MaxItemQuantity)}
	}

	var unitPrice int64
	if targetUser.IsBarteamer {
		unitPrice = item.PriceBarteamer
	} else {
		unitPrice = item.PriceHelfer
	}

	amount := -(unitPrice * int64(quantity))

	var newBalance int64
	err = h.Store.WithTx(ctx, func(tx pgx.Tx) error {
		balance, err := store.GetUserBalanceForUpdate(ctx, tx, targetUser.ID)
		if err != nil {
			return fmt.Errorf("get balance for update: %w", err)
		}

		if settings.HardLimitEnabled && !targetUser.SpendingLimitDisabled {
			if balance+amount <= -settings.HardSpendingLimit {
				return &handler.ValidationError{Message: "Bestellung nicht möglich: Ausgabenlimit erreicht."}
			}
		}

		itemTitle := item.Name
		createdBy := kioskUser.ID
		txn := &model.Transaction{
			UserID:          targetUser.ID,
			Amount:          amount,
			ItemTitle:       &itemTitle,
			UnitPrice:       &unitPrice,
			Quantity:        &quantity,
			Type:            "purchase",
			CreatedByUserID: &createdBy,
		}
		_, err = store.CreateTransaction(ctx, tx, txn)
		if err != nil {
			return err
		}

		newBalance = balance + amount
		return nil
	})
	if err != nil {
		var valErr *handler.ValidationError
		if errors.As(err, &valErr) {
			return valErr
		}
		return fmt.Errorf("kiosk place order: %w", err)
	}

	warning := newBalance < settings.WarningLimit

	h.Renderer.Fragment(w, r, "kiosk-success", map[string]any{
		"Title":       "Bestellung gebucht!",
		"Subtitle":    fmt.Sprintf("%dx %s", quantity, item.Name),
		"UserName":    targetUser.FullName,
		"TotalAmount": -amount,
		"NewBalance":  newBalance,
		"Warning":     warning,
	})

	return nil
}
