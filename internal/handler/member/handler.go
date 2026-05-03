package member

import "github.com/schliz/deckel/internal/handler"

// Handler serves the member-facing endpoints (menu, ordering,
// transaction history, profile).
type Handler struct {
	*handler.Base
}
