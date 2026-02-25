package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/k4-bar/deckel/internal/auth"
	"github.com/k4-bar/deckel/internal/config"
	"github.com/k4-bar/deckel/internal/handler"
	"github.com/k4-bar/deckel/internal/middleware"
	"github.com/k4-bar/deckel/internal/render"
	"github.com/k4-bar/deckel/internal/store"
	"github.com/k4-bar/deckel/migrations"
	"github.com/pressly/goose/v3"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create database connection pool.
	dbpool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to create database pool: %v", err)
	}
	defer dbpool.Close()

	// Run database migrations.
	if err := runMigrations(cfg.DatabaseURL); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize store, renderer, and handler.
	s := store.New(dbpool)
	rndr, err := render.New(cfg.TemplateDir, cfg.DevMode, render.FuncMap())
	if err != nil {
		log.Fatalf("Failed to initialize renderer: %v", err)
	}
	h := &handler.Handler{
		Store:    s,
		Renderer: rndr,
		Config:   cfg,
	}

	// Generate CSRF secret (random 32 bytes).
	csrfSecret := make([]byte, 32)
	if _, err := rand.Read(csrfSecret); err != nil {
		log.Fatalf("Failed to generate CSRF secret: %v", err)
	}

	// Middleware chains.
	base := middleware.Chain(
		middleware.Logging(),
		middleware.Recovery(),
		auth.Middleware(s, cfg.AdminGroup),
	)
	withCSRF := middleware.Chain(
		base,
		middleware.CSRF(csrfSecret),
	)
	mux := http.NewServeMux()

	// Health check - outside middleware (no auth needed).
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		if err := dbpool.Ping(r.Context()); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("database unreachable"))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// Static file server - outside auth middleware.
	staticFS := http.FileServer(http.Dir(cfg.StaticDir))
	mux.Handle("/static/", http.StripPrefix("/static/", staticCacheHandler(staticFS, cfg.DevMode)))

	// Menu page (GET / and GET /menu).
	mux.Handle("GET /{$}", base(h.Wrap(h.MenuPage)))
	mux.Handle("GET /menu", base(h.Wrap(h.MenuPage)))

	// Placeholder routes with base middleware (auth, no CSRF for GET).
	mux.Handle("GET /profile", base(http.HandlerFunc(placeholderHandler("profile"))))
	mux.Handle("GET /transactions", base(http.HandlerFunc(placeholderHandler("transactions"))))

	// Admin routes with CSRF + RequireAdmin.
	adminOnly := func(h http.Handler) http.Handler {
		return withCSRF(auth.RequireAdmin(h))
	}
	mux.Handle("GET /admin/users", adminOnly(http.HandlerFunc(placeholderHandler("admin/users"))))

	srv := &http.Server{
		Addr:    cfg.ListenAddr,
		Handler: mux,
	}

	// Channel to listen for OS signals.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		log.Printf("K4-Bar Deckel starting on %s", cfg.ListenAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Block until a signal is received.
	sig := <-quit
	log.Printf("Received signal %v, shutting down...", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown error: %v", err)
	}

	log.Println("Server stopped gracefully")
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
			}
		}
		next.ServeHTTP(w, r)
	})
}

// placeholderHandler returns a simple handler that responds with the route name.
func placeholderHandler(name string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("placeholder: " + name))
	}
}

func runMigrations(databaseURL string) error {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return err
	}
	defer db.Close()

	goose.SetBaseFS(migrations.FS)

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	if err := goose.Up(db, "."); err != nil {
		return err
	}

	log.Println("Migrations applied successfully")
	return nil
}
