# Deckel

Beverage tab management system for the K4 Bar. Members can order drinks, view their transaction history, and export their data. Admins manage the menu, users, transactions, and settings.

Built with Go, PostgreSQL, HTMX, and DaisyUI. Authentication is handled externally via oauth2-proxy (forwarded headers).

## Quick Start

```sh
docker compose up --build
```

The app is available at `http://localhost:8080`. PostgreSQL starts automatically.

In production, place an [oauth2-proxy](https://oauth2-proxy.github.io/oauth2-proxy/) in front of the app (see `docker-compose.yml` for an example config).

## Development

```sh
# Build CSS (requires Node.js + npm)
npm install && scripts/build-css.sh

# Run the Go server (requires DATABASE_URL)
go build -o deckel ./cmd/server && ./deckel
```
