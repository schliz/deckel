package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/schliz/deckel/internal/config"
	"github.com/schliz/deckel/internal/handler"
	"github.com/schliz/deckel/internal/handler/admin"
	"github.com/schliz/deckel/internal/handler/kiosk"
	"github.com/schliz/deckel/internal/handler/member"
	"github.com/schliz/deckel/internal/handler/shared"
	"github.com/schliz/deckel/internal/render"
	"github.com/schliz/deckel/internal/router"
	"github.com/schliz/deckel/internal/store"
	"github.com/schliz/deckel/migrations"
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

	// Compute content-hashed CSS filename for cache busting.
	cssPath, err := render.CSSHashedPath(cfg.StaticDir, "styles.css")
	if err != nil {
		log.Fatalf("Failed to hash CSS: %v", err)
	}
	log.Printf("Serving CSS as %s", cssPath)

	funcMap := render.FuncMap()
	funcMap["cssFile"] = func() string { return cssPath }
	funcMap["orgName"] = func() string { return cfg.Organization }
	funcMap["appName"] = func() string { return cfg.AppName }

	rndr, err := render.New(cfg.TemplateDir, cfg.DevMode, funcMap)
	if err != nil {
		log.Fatalf("Failed to initialize renderer: %v", err)
	}
	h := &handler.Base{
		Store:    s,
		Renderer: rndr,
		Config:   cfg,
	}
	sharedH := &shared.Handler{Base: h}
	adminH := &admin.Handler{Base: h}
	kioskH := &kiosk.Handler{Base: h}
	memberH := &member.Handler{Base: h}

	// Generate CSRF secret (random 32 bytes).
	csrfSecret := make([]byte, 32)
	if _, err := rand.Read(csrfSecret); err != nil {
		log.Fatalf("Failed to generate CSRF secret: %v", err)
	}

	mux := router.New(router.Deps{
		Base:       h,
		Admin:      adminH,
		Kiosk:      kioskH,
		Member:     memberH,
		Shared:     sharedH,
		Store:      s,
		DBPool:     dbpool,
		Config:     cfg,
		CSRFSecret: csrfSecret,
	})

	srv := &http.Server{
		Addr:    cfg.ListenAddr,
		Handler: mux,
	}

	// Channel to listen for OS signals.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		log.Printf("schliz/deckel starting on %s", cfg.ListenAddr)
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
