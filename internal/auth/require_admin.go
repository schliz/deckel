package auth

import "net/http"

// RequireAdmin returns middleware that blocks non-admin users with 403 Forbidden.
// For HTMX requests it returns a toast error fragment; for normal requests plain text.
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := UserFromContext(r.Context())
		if user == nil || !user.IsAdmin {
			if r.Header.Get("HX-Request") == "true" {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte(adminDeniedToast))
			} else {
				http.Error(w, "Zugriff verweigert", http.StatusForbidden)
			}
			return
		}
		next.ServeHTTP(w, r)
	})
}

const adminDeniedToast = `<div hx-swap-oob="beforeend:#toast-zone">
    <div class="alert alert-error shadow-lg">
        <span>Zugriff verweigert</span>
    </div>
    <script>
        (function() {
            var el = document.currentScript.previousElementSibling;
            setTimeout(function() { el.remove(); }, 4000);
            document.currentScript.remove();
        })();
    </script>
</div>`
