# Guidelines for AI Coding Agents

Use context7 when working on this project to get the required knowledge about the libraries and dependencies used here. Do not use your own knowledge of these dependencies, it is outdated.

## What This Is

Deckel is a beverage tab management system for the K4 Bar. Members order drinks, view transactions, and export data. Admins manage the menu, users, transactions, and settings. Built with Go, PostgreSQL, HTMX, and DaisyUI. Authentication is handled externally via oauth2-proxy forwarded headers (`X-Forwarded-Email`, `X-Forwarded-Groups`, `X-Forwarded-Access-Token`).

## Build & Run Commands

```sh
# Full stack (app + Postgres + oauth2-proxy)
docker compose up --build

# Build and run locally (requires DECKEL_DATABASE_URL env var and running Postgres)
go build -o deckel ./cmd/server && DECKEL_DEV_MODE=true DECKEL_DATABASE_URL="postgres://deckel:deckel@localhost:5432/deckel?sslmode=disable" ./deckel

# Build CSS (requires Node.js, one-time: npm install)
make css

# Rebuild only the app container (force no-cache if templates changed)
podman build --no-cache -t deckel_app . && docker compose up -d --force-recreate app
```

Run `make e2e` for end-to-end tests (see E2E Tests section below). No linters configured yet.

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
| DECKEL_DATABASE_URL | (required) | Postgres connection string |
| DECKEL_LISTEN_ADDR | :8080 | Server bind address |
| DECKEL_DEV_MODE | false | Enables template hot-reload, disables cache headers |
| DECKEL_ADMIN_GROUP | admin | Group name that grants admin access |
| DECKEL_STATIC_DIR | ./static | Path to static assets |
| DECKEL_TEMPLATE_DIR | ./templates | Path to HTML templates |
| DECKEL_ORGANIZATION | K4-Bar | Organization name shown in navbar and page title |
| DECKEL_APP_NAME | Deckel | App name shown in page title |

## Pitfalls

- **OOB swap with `<tr>`**: Browsers strip `<tr>` tags when parsed inside a `<div>` (which HTMX uses for OOB responses). Never use `hx-swap-oob` on table rows. Instead, return the `<tr>` as the primary response and use `hx-target="#row-id"` + `hx-swap="outerHTML"` on the triggering element. See `confirm_toggle_modal.html` and `ToggleActive` handler for the correct pattern.
- **CSRF token**: The token is injected into all HTMX requests via `htmx:configRequest` in `base.html`. Every state-changing endpoint must be behind the CSRF middleware chain.
- **Middleware composition**: Right-to-left — first argument is outermost. Two chains exist (`base` and `withCSRF`). Admin routes add `RequireAdmin` on top. Read `internal/middleware/chain.go` before changing the chain.

## E2E Tests

Playwright e2e tests live in `e2e/`, which is a self-contained directory with its own `package.json` and `playwright.config.ts`. Run `make e2e` to spin up the full Docker test stack (Postgres, Keycloak, oauth2-proxy, coverage-instrumented app), execute tests, and collect Go coverage. `make e2e-up` / `make e2e-down` manage the stack independently. Install test deps with `cd e2e && npm install && npx playwright install --with-deps chromium`. Auth storage state is saved per role in `e2e/.auth/` — the `setup` project drives the real Keycloak login flow. Tests run sequentially against shared DB state. Use cases in `docs/use-cases.md` are the source for test scenarios.

## Node.js Policy

This is a Go project. Node.js is used only as a build tool — the root `package.json` exists solely to download the Tailwind CLI and DaisyUI for CSS generation (`make css`). Do not add application dependencies, scripts, metadata (description, license, author), or test tooling to the root `package.json`. Build tasks go in the `Makefile`, not in npm scripts. The `e2e/` directory has its own isolated `package.json` for Playwright.

## Use Cases (`docs/use-cases.md`)

Concise acceptance-level use cases used to derive e2e tests. Schema:

```markdown
## <Concern>
- <Name>: <Akteur> does <Interaktion>. <Akzeptanzkriterium>.
```

When adding or changing user-facing behaviour (new route, new UI interaction, changed business rule), update the use cases file to match. Keep entries short — one sentence for the interaction, one for the expected outcome. Do not duplicate, reorder, or reformat existing entries.
