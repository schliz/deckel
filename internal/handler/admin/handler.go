package admin

import (
	"sync"

	"github.com/schliz/deckel/internal/handler"
)

// Handler serves all /admin/* endpoints.
type Handler struct {
	*handler.Base

	// MenuBatchSessions holds in-flight batch-import previews per
	// CSRF token. A sync.Map's zero value is safe to use directly.
	MenuBatchSessions sync.Map
}
