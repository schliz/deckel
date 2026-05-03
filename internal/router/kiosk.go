package router

import (
	"net/http"

	"github.com/schliz/deckel/internal/auth"
	"github.com/schliz/deckel/internal/handler/kiosk"
)

// RegisterKiosk wires the /kiosk/* routes. All routes require the kiosk
// role (auth.RequireKiosk) on top of CSRF + base middleware.
func RegisterKiosk(mux *http.ServeMux, h *kiosk.Handler, withCSRF func(http.Handler) http.Handler) {
	kioskOnly := func(handler http.Handler) http.Handler {
		return withCSRF(auth.RequireKiosk(handler))
	}

	mux.Handle("GET /kiosk", kioskOnly(h.Wrap(h.KioskMenuPage)))
	mux.Handle("GET /kiosk/items/{id}/users", kioskOnly(h.Wrap(h.KioskUserSelect)))
	mux.Handle("GET /kiosk/items/{id}/confirm/{uid}", kioskOnly(h.Wrap(h.KioskConfirm)))
	mux.Handle("POST /kiosk/order", kioskOnly(h.Wrap(h.KioskPlaceOrder)))
	mux.Handle("GET /kiosk/history", kioskOnly(h.Wrap(h.KioskHistory)))
	mux.Handle("GET /kiosk/transactions/{id}/cancel", kioskOnly(h.Wrap(h.KioskCancelModal)))
	mux.Handle("POST /kiosk/transactions/{id}/cancel", kioskOnly(h.Wrap(h.KioskCancelTransaction)))
}
