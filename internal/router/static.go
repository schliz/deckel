package router

import (
	"net/http"
	"path/filepath"
	"strings"
)

// registerStatic wires the routes that must bypass auth: the health check,
// the static file server (with cache headers), and the PWA assets served
// from the URL root.
func registerStatic(mux *http.ServeMux, d Deps, base func(http.Handler) http.Handler) {
	// Health check - outside middleware (no auth needed).
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		if err := d.DBPool.Ping(r.Context()); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("database unreachable"))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// Static file server - outside auth middleware.
	staticFS := http.FileServer(http.Dir(d.Config.StaticDir))
	mux.Handle("/static/", http.StripPrefix("/static/", staticCacheHandler(staticFS, d.Config.DevMode)))

	// PWA assets served from root - must bypass auth (see OAUTH2_PROXY_SKIP_AUTH_REGEX).
	mux.HandleFunc("GET /sw.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache")
		http.ServeFile(w, r, filepath.Join(d.Config.StaticDir, "js", "sw.js"))
	})
	mux.HandleFunc("GET /manifest.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=86400")
		http.ServeFile(w, r, filepath.Join(d.Config.StaticDir, "manifest.json"))
	})
	mux.HandleFunc("GET /favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=86400")
		http.ServeFile(w, r, filepath.Join(d.Config.StaticDir, "icons", "favicon.ico"))
	})
}

// staticCacheHandler wraps a file server handler to add cache headers based on file type.
// In dev mode, cache headers are set to no-cache.
func staticCacheHandler(next http.Handler, devMode bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if devMode {
			w.Header().Set("Cache-Control", "no-cache")
		} else {
			path := r.URL.Path
			switch {
			case strings.HasSuffix(path, ".css"):
				w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			case strings.HasSuffix(path, ".js"):
				w.Header().Set("Cache-Control", "public, max-age=86400")
			case strings.HasSuffix(path, ".png"), strings.HasSuffix(path, ".json"):
				w.Header().Set("Cache-Control", "public, max-age=86400")
			}
		}
		next.ServeHTTP(w, r)
	})
}
