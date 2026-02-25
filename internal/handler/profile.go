package handler

import (
	"net/http"

	"github.com/k4-bar/deckel/internal/auth"
	"github.com/k4-bar/deckel/internal/middleware"
)

// ProfilePageData is the view model for the profile page.
type ProfilePageData struct {
	User       *auth.RequestUser
	CSRFToken  string
	ActivePage string
}

// ProfilePage renders the user profile page.
func (h *Handler) ProfilePage(w http.ResponseWriter, r *http.Request) error {
	user := auth.UserFromContext(r.Context())

	data := ProfilePageData{
		User:       user,
		CSRFToken:  middleware.CSRFTokenFromContext(r.Context()),
		ActivePage: "profile",
	}

	h.Renderer.Page(w, r, "profile", data)
	return nil
}
