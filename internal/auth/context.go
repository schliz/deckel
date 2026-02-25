package auth

import "context"

// RequestUser represents the authenticated user extracted from request headers.
type RequestUser struct {
	Email       string
	FullName    string
	GivenName   string
	FamilyName  string
	Groups      []string
	IsAdmin     bool
	ID          int64
	Balance     int64
	IsBarteamer bool
	IsActive    bool
}

type contextKey struct{}

// UserFromContext retrieves the RequestUser from the context.
// Returns nil if no user is stored in the context.
func UserFromContext(ctx context.Context) *RequestUser {
	u, _ := ctx.Value(contextKey{}).(*RequestUser)
	return u
}

// contextWithUser stores the RequestUser in the context.
func contextWithUser(ctx context.Context, user *RequestUser) context.Context {
	return context.WithValue(ctx, contextKey{}, user)
}
