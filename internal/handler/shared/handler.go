package shared

import "github.com/schliz/deckel/internal/handler"

// Handler exposes endpoints that don't belong to one specific domain
// (e.g. the lazy-loaded header stats fragment used on every page).
type Handler struct {
	*handler.Base
}
