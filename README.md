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

> [!TIP]
> While originally written for ai coding agents, the [CLAUDE.md](CLAUDE.md) file provides a brief overview that can be beneficial to human contributors as well.

```sh
# Build CSS (requires Node.js + npm, one-time: npm install)
make css

# Run the Go server (requires DECKEL_DATABASE_URL)
go build -o deckel ./cmd/server && ./deckel
```

For local testing with oauth2-proxy reachable only via HTTP, use the following `docker-compose.override.yml`:

```yaml
services:
  oauth2_proxy:
    ports:
      - "4180:4180"
    environment:
      OAUTH2_PROXY_COOKIE_SECURE: "false"
```
