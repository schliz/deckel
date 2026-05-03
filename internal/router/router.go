package router

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/schliz/deckel/internal/auth"
	"github.com/schliz/deckel/internal/config"
	"github.com/schliz/deckel/internal/handler"
	"github.com/schliz/deckel/internal/handler/admin"
	"github.com/schliz/deckel/internal/handler/kiosk"
	"github.com/schliz/deckel/internal/handler/member"
	"github.com/schliz/deckel/internal/handler/shared"
	"github.com/schliz/deckel/internal/middleware"
	"github.com/schliz/deckel/internal/store"
)

// Deps bundles everything the router needs to wire HTTP routes.
type Deps struct {
	Base       *handler.Base
	Admin      *admin.Handler
	Kiosk      *kiosk.Handler
	Member     *member.Handler
	Shared     *shared.Handler
	Store      *store.Store
	DBPool     *pgxpool.Pool
	Config     *config.Config
	CSRFSecret []byte
}

// New builds the application's HTTP mux with middleware chains and all
// route registrations. The returned handler is ready to serve.
func New(d Deps) http.Handler {
	base := middleware.Chain(
		middleware.Logging(),
		middleware.Recovery(),
		auth.Middleware(d.Store, d.Config.AdminGroup, d.Config.KioskGroup, d.Config.Organization, d.Config.AppName),
	)
	withCSRF := middleware.Chain(base, middleware.CSRF(d.CSRFSecret))

	mux := http.NewServeMux()

	registerStatic(mux, d, base)
	registerShared(mux, d.Base, d.Shared, base)
	RegisterMember(mux, d.Member, withCSRF)
	RegisterKiosk(mux, d.Kiosk, withCSRF)
	RegisterAdmin(mux, d.Admin, withCSRF)

	return mux
}
