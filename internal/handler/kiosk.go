package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/schliz/deckel/internal/auth"
	"github.com/schliz/deckel/internal/middleware"
	"github.com/schliz/deckel/internal/model"
	"github.com/schliz/deckel/internal/store"
)

// KioskMenuPage renders the kiosk item selection grid.
func (h *Base) KioskMenuPage(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	user := auth.UserFromContext(ctx)
	db := h.Store.DB()

	cats, err := store.ListCategories(ctx, db)
	if err != nil {
		return fmt.Errorf("kiosk menu: list categories: %w", err)
	}

	var categories []CategoryWithItems
	for _, cat := range cats {
		items, err := store.ListItemsByCategory(ctx, db, cat.ID)
		if err != nil {
			return fmt.Errorf("kiosk menu: list items for category %d: %w", cat.ID, err)
		}
		if len(items) > 0 {
			categories = append(categories, CategoryWithItems{
				Category: cat,
				Items:    items,
			})
		}
	}

	data := map[string]any{
		"Categories": categories,
		"CSRFToken":  middleware.CSRFTokenFromContext(ctx),
		"User":       user,
	}

	h.Renderer.Page(w, r, "kiosk_menu", data)
	return nil
}

// KioskUserSelect renders the user selection page for a kiosk order.
func (h *Base) KioskUserSelect(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	user := auth.UserFromContext(ctx)
	db := h.Store.DB()

	idStr := r.PathValue("id")
	itemID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return &NotFoundError{Message: "Artikel nicht gefunden"}
	}

	item, err := store.GetItem(ctx, db, itemID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return &NotFoundError{Message: "Artikel nicht gefunden"}
		}
		return fmt.Errorf("kiosk user select: get item: %w", err)
	}
	if item.DeletedAt != nil {
		return &NotFoundError{Message: "Artikel nicht gefunden"}
	}

	users, err := store.ListActiveUsersWithBalance(ctx, db)
	if err != nil {
		return fmt.Errorf("kiosk user select: list users: %w", err)
	}

	// Filter out the kiosk user itself.
	var filtered []model.UserWithBalance
	for _, u := range users {
		if u.ID != user.ID {
			filtered = append(filtered, u)
		}
	}

	settings, err := store.GetSettings(ctx, db)
	if err != nil {
		return fmt.Errorf("kiosk user select: get settings: %w", err)
	}

	data := map[string]any{
		"Item":      item,
		"Users":     filtered,
		"Settings":  settings,
		"CSRFToken": middleware.CSRFTokenFromContext(ctx),
	}

	h.Renderer.Page(w, r, "kiosk_users", data)
	return nil
}

// KioskConfirm renders the order confirmation page for a kiosk order.
func (h *Base) KioskConfirm(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	db := h.Store.DB()

	itemID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		return &NotFoundError{Message: "Artikel nicht gefunden"}
	}

	userID, err := strconv.ParseInt(r.PathValue("uid"), 10, 64)
	if err != nil {
		return &NotFoundError{Message: "Benutzer nicht gefunden"}
	}

	item, err := store.GetItem(ctx, db, itemID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return &NotFoundError{Message: "Artikel nicht gefunden"}
		}
		return fmt.Errorf("kiosk confirm: get item: %w", err)
	}
	if item.DeletedAt != nil {
		return &NotFoundError{Message: "Artikel nicht gefunden"}
	}

	targetUser, err := store.GetUserWithBalance(ctx, db, userID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return &NotFoundError{Message: "Benutzer nicht gefunden"}
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

	h.Renderer.Page(w, r, "kiosk_confirm", data)
	return nil
}

// KioskPlaceOrder processes a kiosk order submission.
func (h *Base) KioskPlaceOrder(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	kioskUser := auth.UserFromContext(ctx)
	db := h.Store.DB()

	itemID, err := strconv.ParseInt(r.FormValue("item_id"), 10, 64)
	if err != nil {
		return &ValidationError{Message: "Ungültige Artikel-ID"}
	}

	userID, err := strconv.ParseInt(r.FormValue("user_id"), 10, 64)
	if err != nil {
		return &ValidationError{Message: "Ungültige Benutzer-ID"}
	}

	quantity, err := strconv.Atoi(r.FormValue("quantity"))
	if err != nil || quantity < 1 {
		return &ValidationError{Message: "Ungültige Menge"}
	}

	item, err := store.GetItem(ctx, db, itemID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return &NotFoundError{Message: "Artikel nicht gefunden"}
		}
		return fmt.Errorf("kiosk place order: get item: %w", err)
	}
	if item.DeletedAt != nil {
		return &NotFoundError{Message: "Artikel nicht gefunden"}
	}

	targetUser, err := store.GetUserWithBalance(ctx, db, userID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return &NotFoundError{Message: "Benutzer nicht gefunden"}
		}
		return fmt.Errorf("kiosk place order: get user: %w", err)
	}

	if !targetUser.IsActive {
		return &ValidationError{Message: "Benutzer ist deaktiviert"}
	}

	settings, err := store.GetSettings(ctx, db)
	if err != nil {
		return fmt.Errorf("kiosk place order: get settings: %w", err)
	}

	if quantity > settings.MaxItemQuantity {
		return &ValidationError{Message: fmt.Sprintf("Maximal %d Stück erlaubt", settings.MaxItemQuantity)}
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
				return &ValidationError{Message: "Bestellung nicht möglich: Ausgabenlimit erreicht."}
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
		var valErr *ValidationError
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

// KioskHistory renders the recent kiosk transaction history.
func (h *Base) KioskHistory(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	user := auth.UserFromContext(ctx)
	db := h.Store.DB()

	txns, err := store.ListTransactionsByCreator(ctx, db, user.ID, 20)
	if err != nil {
		return fmt.Errorf("kiosk history: list transactions: %w", err)
	}

	settings, err := store.GetSettings(ctx, db)
	if err != nil {
		return fmt.Errorf("kiosk history: get settings: %w", err)
	}

	data := map[string]any{
		"Transactions": txns,
		"Settings":     settings,
		"CSRFToken":    middleware.CSRFTokenFromContext(ctx),
		"KioskUserID":  user.ID,
	}

	h.Renderer.Page(w, r, "kiosk_history", data)
	return nil
}

// KioskCancelModal renders the cancel confirmation modal for a kiosk transaction.
func (h *Base) KioskCancelModal(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	user := auth.UserFromContext(ctx)
	db := h.Store.DB()

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		return &NotFoundError{Message: "Transaktion nicht gefunden"}
	}

	txn, err := store.GetTransaction(ctx, db, id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return &NotFoundError{Message: "Transaktion nicht gefunden"}
		}
		return fmt.Errorf("kiosk cancel modal: get transaction: %w", err)
	}

	// Verify this transaction was created by the kiosk user.
	if txn.CreatedByUserID == nil || *txn.CreatedByUserID != user.ID {
		return &ForbiddenError{Message: "Zugriff verweigert"}
	}

	if txn.CancelledAt != nil {
		return &ValidationError{Message: "Transaktion wurde bereits storniert"}
	}

	if txn.Type == "cancellation" {
		return &ValidationError{Message: "Stornobuchungen können nicht storniert werden"}
	}

	settings, err := store.GetSettings(ctx, db)
	if err != nil {
		return fmt.Errorf("kiosk cancel modal: get settings: %w", err)
	}

	if time.Since(txn.CreatedAt) > time.Duration(settings.CancellationMinutes)*time.Minute {
		return &ValidationError{Message: "Stornierungsfenster abgelaufen"}
	}

	// We need the user name for the modal display - fetch it.
	targetUser, err := store.GetUserWithBalance(ctx, db, txn.UserID)
	if err != nil {
		return fmt.Errorf("kiosk cancel modal: get target user: %w", err)
	}

	data := map[string]any{
		"Transaction": struct {
			*model.Transaction
			UserName string
		}{txn, targetUser.FullName},
		"CSRFToken": middleware.CSRFTokenFromContext(ctx),
	}

	h.Renderer.Fragment(w, r, "kiosk-cancel-modal", data)
	return nil
}

// KioskCancelTransaction processes the cancellation of a kiosk-created transaction.
func (h *Base) KioskCancelTransaction(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	user := auth.UserFromContext(ctx)
	db := h.Store.DB()

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		return &NotFoundError{Message: "Transaktion nicht gefunden"}
	}

	txn, err := store.GetTransaction(ctx, db, id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return &NotFoundError{Message: "Transaktion nicht gefunden"}
		}
		return fmt.Errorf("kiosk cancel: get transaction: %w", err)
	}

	if txn.CreatedByUserID == nil || *txn.CreatedByUserID != user.ID {
		return &ForbiddenError{Message: "Zugriff verweigert"}
	}

	if txn.CancelledAt != nil {
		return &ValidationError{Message: "Transaktion wurde bereits storniert"}
	}

	if txn.Type == "cancellation" {
		return &ValidationError{Message: "Stornobuchungen können nicht storniert werden"}
	}

	settings, err := store.GetSettings(ctx, db)
	if err != nil {
		return fmt.Errorf("kiosk cancel: get settings: %w", err)
	}

	if time.Since(txn.CreatedAt) > time.Duration(settings.CancellationMinutes)*time.Minute {
		return &ValidationError{Message: "Stornierungsfenster abgelaufen"}
	}

	err = h.Store.WithTx(ctx, func(tx pgx.Tx) error {
		return store.CancelTransaction(ctx, tx, id, user.ID)
	})
	if err != nil {
		return fmt.Errorf("kiosk cancel: %w", err)
	}

	// Re-render the kiosk history list.
	txns, err := store.ListTransactionsByCreator(ctx, db, user.ID, 20)
	if err != nil {
		return fmt.Errorf("kiosk cancel: list transactions: %w", err)
	}

	// Render toast + refreshed history list + close modal.
	h.Renderer.Fragment(w, r, "toast", map[string]string{
		"Type":    "success",
		"Message": "Transaktion storniert!",
	})

	h.Renderer.AppendOOB(w, "kiosk-history-list", map[string]any{
		"Transactions": txns,
		"Settings":     settings,
		"KioskUserID":  user.ID,
		"OOB":          true,
	})

	w.Write([]byte(`<div id="modal" hx-swap-oob="innerHTML" style="display:none"></div>`))

	return nil
}
