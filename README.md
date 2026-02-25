# Deckel

Beverage tab management system for the K4 Bar. Members can order drinks, view their transaction history, and export their data. Admins manage the menu, users, transactions, and settings.

Built with Go, PostgreSQL, HTMX, and DaisyUI. Authentication is handled externally via oauth2-proxy (forwarded headers).

## Quick Start

```sh
cp .env.example .env
docker compose up --build
```

After connecting an OIDC IdP the app is available at `http://localhost:4180`. PostgreSQL starts automatically.

## Development

```sh
# Build CSS (requires Node.js + npm)
npm install && scripts/build-css.sh

# Run the Go server (requires DATABASE_URL)
go build -o deckel ./cmd/server && ./deckel
```
