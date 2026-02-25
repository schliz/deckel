package middleware

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"net/http"
)

type csrfContextKey struct{}

// CSRFTokenFromContext retrieves the CSRF token from the request context.
func CSRFTokenFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(csrfContextKey{}).(string); ok {
		return v
	}
	return ""
}

// CSRF returns a middleware that provides CSRF protection.
// It generates a random token stored in an HttpOnly cookie and validates
// the X-CSRF-Token header on mutating requests (POST, PUT, DELETE, PATCH).
// The secret is used to HMAC-sign the token for tamper detection.
func CSRF(secret []byte) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := getOrCreateToken(w, r, secret)

			// Store token in context for templates.
			ctx := context.WithValue(r.Context(), csrfContextKey{}, token)
			r = r.WithContext(ctx)

			// On mutating methods, validate the token.
			switch r.Method {
			case http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch:
				headerToken := r.Header.Get("X-CSRF-Token")
				if headerToken == "" || headerToken != token {
					slog.Warn("CSRF token mismatch",
						"method", r.Method,
						"path", r.URL.Path,
					)
					http.Error(w, "Forbidden: CSRF token mismatch", http.StatusForbidden)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// getOrCreateToken reads the CSRF token from the cookie. If absent or invalid,
// it generates a new token and sets the cookie.
func getOrCreateToken(w http.ResponseWriter, r *http.Request, secret []byte) string {
	cookie, err := r.Cookie("_csrf")
	if err == nil {
		// Validate the cookie value: it should be token.signature
		token, ok := validateSignedToken(cookie.Value, secret)
		if ok {
			return token
		}
	}

	// Generate new token.
	token := generateToken()
	signed := signToken(token, secret)

	http.SetCookie(w, &http.Cookie{
		Name:     "_csrf",
		Value:    signed,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	return token
}

// generateToken creates a 32-byte random hex-encoded token.
func generateToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic("csrf: failed to generate random token: " + err.Error())
	}
	return hex.EncodeToString(b)
}

// signToken returns "token.hmac_signature" to prevent cookie tampering.
func signToken(token string, secret []byte) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(token))
	sig := hex.EncodeToString(mac.Sum(nil))
	return token + "." + sig
}

// validateSignedToken splits the signed value, verifies the HMAC, and returns the token.
func validateSignedToken(signed string, secret []byte) (string, bool) {
	// Find the last dot separator.
	dotIdx := -1
	for i := len(signed) - 1; i >= 0; i-- {
		if signed[i] == '.' {
			dotIdx = i
			break
		}
	}
	if dotIdx <= 0 || dotIdx >= len(signed)-1 {
		return "", false
	}

	token := signed[:dotIdx]
	sig := signed[dotIdx+1:]

	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(token))
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(sig), []byte(expectedSig)) {
		return "", false
	}

	return token, true
}
