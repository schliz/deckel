package middleware

import "net/http"

// Middleware is a function that wraps an http.Handler with additional behavior.
type Middleware func(http.Handler) http.Handler

// Chain composes middlewares right-to-left so the first argument is the
// outermost middleware (executed first on a request).
// Usage: Chain(logging, recovery, auth).Then(finalHandler)
func Chain(mws ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(mws) - 1; i >= 0; i-- {
			next = mws[i](next)
		}
		return next
	}
}
