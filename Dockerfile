# Stage 1: Build CSS with Tailwind + DaisyUI
FROM node:22-alpine AS tailwind
WORKDIR /src
COPY package.json package-lock.json* ./
RUN npm ci
COPY static/css/input.css ./static/css/input.css
COPY templates/ ./templates/
COPY internal/ ./internal/
RUN npx @tailwindcss/cli -i static/css/input.css -o static/css/styles.css --minify

# Stage 2: Build Go binary
FROM golang:1.25-alpine AS builder
ARG TARGETARCH
# Defaults produce a stripped release build; for a coverage-instrumented build
# override with --build-arg GO_LDFLAGS="" --build-arg GO_COVER=-cover
ARG GO_LDFLAGS="-s -w"
ARG GO_COVER=""
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=tailwind /src/static/css/styles.css ./static/css/styles.css
RUN CGO_ENABLED=0 GOARCH=${TARGETARCH} go build -ldflags="${GO_LDFLAGS}" ${GO_COVER} -o /bin/deckel ./cmd/server

# Stage 3: Minimal runtime
FROM alpine:3.22
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /bin/deckel /usr/local/bin/deckel
COPY --from=builder /src/static/ /app/static/
COPY --from=builder /src/templates/ /app/templates/
COPY --from=builder /src/migrations/ /app/migrations/
WORKDIR /app
EXPOSE 8080
ENTRYPOINT ["deckel"]
