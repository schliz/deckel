# CLAUDE.md

Use context7 when working on this project to get the required knowledge about the libraries and dependencies used here. Do not use your own knowledge of these dependencies, it is outdated.

## What This Is

Deckel is a beverage tab management system for the K4 Bar. Members order drinks, view transactions, and export data. Admins manage the menu, users, transactions, and settings. Built with Go, PostgreSQL, HTMX, and DaisyUI. Authentication is handled externally via oauth2-proxy forwarded headers (`X-Forwarded-Email`, `X-Forwarded-Groups`, `X-Forwarded-Access-Token`).

## Build & Run Commands

```sh
# Full stack (app + Postgres + oauth2-proxy)
docker compose up --build

# Build and run locally (requires DATABASE_URL env var and running Postgres)
go build -o deckel ./cmd/server && DEV_MODE=true DATABASE_URL="postgres://deckel:deckel@localhost:5432/deckel?sslmode=disable" ./deckel

# Build CSS (requires Node.js)
npm install && scripts/build-css.sh

# Rebuild only the app container (force no-cache if templates changed)
podman build --no-cache -t deckel_app . && docker compose up -d --force-recreate app
```

There are no tests or linters configured yet.

## Architecture

Server-rendered Go app. Handlers return errors, a central `Wrap()` function maps them to HTTP responses. Store methods accept a `DBTX` interface so they work with both the connection pool and transactions. Templates are HTMX-driven with OOB swaps for partial updates. Explore the codebase to understand specifics — the layout is conventional:

- `cmd/server/` — entrypoint, routing
- `internal/handler/` — HTTP handlers
- `internal/store/` — database queries
- `internal/render/` — template rendering
- `internal/auth/` — authentication middleware
- `internal/middleware/` — middleware chain
- `migrations/` — Goose migrations (embedded, auto-applied at startup)
- `templates/` — HTML templates (`layouts/`, `pages/`, `partials/`, `components/`)
- `static/` — CSS and assets

All monetary values are stored as cents (bigint). Items use soft-delete (`deleted_at`).

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| DATABASE_URL | (required) | Postgres connection string |
| LISTEN_ADDR | :8080 | Server bind address |
| DEV_MODE | false | Enables template hot-reload, disables cache headers |
| ADMIN_GROUP | admin | Group name that grants admin access |
| STATIC_DIR | ./static | Path to static assets |
| TEMPLATE_DIR | ./templates | Path to HTML templates |

## Pitfalls

- **OOB swap with `<tr>`**: Browsers strip `<tr>` tags when parsed inside a `<div>` (which HTMX uses for OOB responses). Never use `hx-swap-oob` on table rows. Instead, return the `<tr>` as the primary response and use `hx-target="#row-id"` + `hx-swap="outerHTML"` on the triggering element. See `confirm_toggle_modal.html` and `ToggleActive` handler for the correct pattern.
- **CSRF token**: The token is injected into all HTMX requests via `htmx:configRequest` in `base.html`. Every state-changing endpoint must be behind the CSRF middleware chain.
- **Middleware composition**: Right-to-left — first argument is outermost. Two chains exist (`base` and `withCSRF`). Admin routes add `RequireAdmin` on top. Read `internal/middleware/chain.go` before changing the chain.

## Use Cases (`docs/use-cases.md`)

Concise acceptance-level use cases used to derive e2e tests. Schema:

```markdown
## <Concern>
- <Name>: <Akteur> does <Interaktion>. <Akzeptanzkriterium>.
```

When adding or changing user-facing behaviour (new route, new UI interaction, changed business rule), update the use cases file to match. Keep entries short — one sentence for the interaction, one for the expected outcome. Do not duplicate, reorder, or reformat existing entries.
