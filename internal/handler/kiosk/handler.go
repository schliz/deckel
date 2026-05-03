package kiosk

import "github.com/schliz/deckel/internal/handler"

// Handler serves all /kiosk/* endpoints (used by the in-bar tablet).
type Handler struct {
	*handler.Base
}
