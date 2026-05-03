package router

import (
	"net/http"

	"github.com/schliz/deckel/internal/auth"
	"github.com/schliz/deckel/internal/handler/member"
)

// RegisterMember wires the member-facing routes (menu, profile, transactions).
// The root route also handles the kiosk-redirect: kiosk users hitting "/"
// are redirected to "/kiosk" instead of seeing the member menu.
func RegisterMember(mux *http.ServeMux, h *member.Handler, withCSRF func(http.Handler) http.Handler) {
	// Menu page (GET / and GET /menu).
	mux.Handle("GET /{$}", withCSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if user := auth.UserFromContext(r.Context()); user != nil && user.IsKiosk {
			http.Redirect(w, r, "/kiosk", http.StatusFound)
			return
		}
		h.Wrap(h.MenuPage)(w, r)
	})))
	mux.Handle("GET /menu", withCSRF(h.Wrap(h.MenuPage)))

	// Order modal (GET /menu/items/{id}/order).
	mux.Handle("GET /menu/items/{id}/order", withCSRF(h.Wrap(h.OrderModal)))

	// Place order (POST /menu/order).
	mux.Handle("POST /menu/order", withCSRF(h.Wrap(h.PlaceOrder)))

	// Profile + transactions (auth + CSRF).
	mux.Handle("GET /profile", withCSRF(h.Wrap(h.ProfilePage)))
	mux.Handle("POST /profile/export", withCSRF(h.Wrap(h.ExportData)))
	mux.Handle("GET /transactions", withCSRF(h.Wrap(h.TransactionHistory)))
	mux.Handle("GET /transactions/custom", withCSRF(h.Wrap(h.CustomTransactionModal)))
	mux.Handle("POST /transactions/custom", withCSRF(h.Wrap(h.CreateCustomTransaction)))
	mux.Handle("GET /transactions/{id}/cancel", withCSRF(h.Wrap(h.CancelModal)))
	mux.Handle("POST /transactions/{id}/cancel", withCSRF(h.Wrap(h.CancelTransaction)))
}
