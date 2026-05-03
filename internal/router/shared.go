package router

import (
	"net/http"

	"github.com/schliz/deckel/internal/handler"
	"github.com/schliz/deckel/internal/handler/shared"
)

// registerShared wires routes that don't belong to any single domain:
// the lazy-loaded header stats endpoint and the catch-all 404 handler.
func registerShared(mux *http.ServeMux, base *handler.Base, sharedH *shared.Handler, baseMW func(http.Handler) http.Handler) {
	// Header stats (lazy-loaded on page init).
	mux.Handle("GET /header-stats", baseMW(sharedH.Wrap(sharedH.HeaderStats)))

	// Catch-all: styled 404 for unmatched routes.
	mux.Handle("/", baseMW(base.NotFoundHandler()))
}
