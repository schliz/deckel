package auth

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/k4-bar/deckel/internal/model"
	"github.com/k4-bar/deckel/internal/store"
)

// Middleware returns HTTP middleware that extracts the authenticated user from
// oauth2-proxy forwarded headers, enriches with DB data, and stores in context.
func Middleware(s *store.Store, adminGroup string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			email := r.Header.Get("X-Forwarded-Email")
			if email == "" {
				if r.Header.Get("HX-Request") == "true" {
					w.Header().Set("HX-Redirect", "/")
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Parse groups from header
			var groups []string
			if gh := r.Header.Get("X-Forwarded-Groups"); gh != "" {
				for _, g := range strings.Split(gh, ",") {
					if trimmed := strings.TrimSpace(g); trimmed != "" {
						groups = append(groups, trimmed)
					}
				}
			}

			isAdmin := false
			for _, g := range groups {
				if g == adminGroup {
					isAdmin = true
					break
				}
			}

			// Extract name claims from JWT access token
			var fullName, givenName, familyName string
			if token := r.Header.Get("X-Forwarded-Access-Token"); token != "" {
				fullName, givenName, familyName = parseJWTNames(token)
			}

			ctx := r.Context()
			db := s.DB()

			// Look up user; auto-create if not found
			user, err := store.GetUserByEmail(ctx, db, email)
			if err != nil {
				if !errors.Is(err, store.ErrNotFound) {
					slog.Error("auth: failed to get user", "email", email, "error", err)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return
				}
				// Auto-create user
				user = &model.User{
					Email:      email,
					FullName:   fullName,
					GivenName:  givenName,
					FamilyName: familyName,
					IsAdmin:    isAdmin,
					IsActive:   true,
				}
				user, err = store.CreateUser(ctx, db, user)
				if err != nil {
					slog.Error("auth: failed to create user", "email", email, "error", err)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return
				}
			} else {
				// Update profile on every request (upsert pattern)
				if err := store.UpdateUserProfile(ctx, db, user.ID, fullName, givenName, familyName, isAdmin); err != nil {
					slog.Error("auth: failed to update user profile", "email", email, "error", err)
					// Non-fatal: continue with existing data
				} else {
					user.FullName = fullName
					user.GivenName = givenName
					user.FamilyName = familyName
					user.IsAdmin = isAdmin
				}
			}

			// Get balance
			balance, err := store.GetUserBalance(ctx, db, user.ID)
			if err != nil {
				slog.Error("auth: failed to get user balance", "email", email, "error", err)
				// Non-fatal: balance stays 0
			}

			reqUser := &RequestUser{
				Email:                 user.Email,
				FullName:              user.FullName,
				GivenName:             user.GivenName,
				FamilyName:            user.FamilyName,
				Groups:                groups,
				IsAdmin:               user.IsAdmin,
				ID:                    user.ID,
				Balance:               balance,
				IsBarteamer:           user.IsBarteamer,
				IsActive:              user.IsActive,
				SpendingLimitDisabled: user.SpendingLimitDisabled,
				CreatedAt:             user.CreatedAt,
			}

			// Block deactivated users
			if !reqUser.IsActive {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte(deactivatedPageHTML))
				return
			}

			next.ServeHTTP(w, r.WithContext(contextWithUser(ctx, reqUser)))
		})
	}
}

// deactivatedPageHTML is the full HTML page shown to deactivated users.
const deactivatedPageHTML = `<!DOCTYPE html>
<html lang="de" data-theme="night">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Konto deaktiviert - K4-Bar Deckel</title>
<link rel="stylesheet" href="/static/css/styles.css">
</head>
<body class="min-h-screen flex items-center justify-center bg-base-200">
<div class="card bg-base-100 shadow-xl max-w-md mx-auto">
<div class="card-body text-center">
<h1 class="card-title justify-center text-2xl">Konto deaktiviert</h1>
<p class="py-4">Dein Konto ist deaktiviert. Bitte wende dich an einen Admin.</p>
</div>
</div>
</body>
</html>`	

// parseJWTNames extracts name, given_name, family_name from a JWT's payload.
// It base64-decodes the payload segment without verifying the signature.
func parseJWTNames(token string) (fullName, givenName, familyName string) {
	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		return
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return
	}

	var claims struct {
		Name       string `json:"name"`
		GivenName  string `json:"given_name"`
		FamilyName string `json:"family_name"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return
	}

	fullName = strings.TrimSpace(claims.Name)
	if fullName == "" {
		fullName = strings.TrimSpace(claims.GivenName + " " + claims.FamilyName)
	}

	return fullName, claims.GivenName, claims.FamilyName
}
