# Stage 1: Build CSS with Tailwind + DaisyUI
FROM node:lts-alpine AS tailwind
WORKDIR /src
COPY package.json package-lock.json* ./
RUN npm ci
COPY static/css/input.css ./static/css/input.css
COPY templates/ ./templates/
RUN npx @tailwindcss/cli -i static/css/input.css -o static/css/styles.css --minify

# Stage 2: Build Go binary
FROM golang:alpine AS builder
ARG TARGETARCH
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=tailwind /src/static/css/styles.css ./static/css/styles.css
RUN CGO_ENABLED=0 GOARCH=${TARGETARCH} go build -ldflags='-s -w' -o /bin/deckel ./cmd/server

# Stage 3: Minimal runtime
FROM alpine:latest
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /bin/deckel /usr/local/bin/deckel
COPY --from=builder /src/static/ /app/static/
COPY --from=builder /src/templates/ /app/templates/
COPY --from=builder /src/migrations/ /app/migrations/
WORKDIR /app
EXPOSE 8080
ENTRYPOINT ["deckel"]
