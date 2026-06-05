FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o luminous ./cmd/server/

FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder --chown=65534:65534 /build/luminous .

# ── All configuration via environment variables (LUMINOUS_ prefix) ──
# Server:
#   LUMINOUS_SERVER_PORT=8080
#   LUMINOUS_SERVER_MODE=release
#   LUMINOUS_SERVER_CORS_ORIGIN=https://myapp.example.com
#   LUMINOUS_SERVER_TLS_CERT=/path/to/cert.pem
#   LUMINOUS_SERVER_TLS_KEY=/path/to/key.pem
#   LUMINOUS_SERVER_TRUSTED_PROXIES=10.0.0.0/8,172.16.0.0/12
# Auth:
#   LUMINOUS_AUTH_ADMIN_TOKEN=<your-secret-token>
# Database:
#   LUMINOUS_DATABASE_DSN=postgresql://user:password@host:port/dbname
#   LUMINOUS_DATABASE_POOL_MAX_CONNS=20
#   LUMINOUS_DATABASE_POOL_MIN_CONNS=5
# Rate limit:
#   LUMINOUS_RATE_LIMIT_RATE=10
#   LUMINOUS_RATE_LIMIT_BURST=30
# Release proxy:
#   LUMINOUS_RELEASE_API_URL=<full-override-url>
#   LUMINOUS_RELEASE_APP_UUID=<app-uuid>
#   LUMINOUS_RELEASE_CHANNEL_ID=<channel-id>

ENV LUMINOUS_SERVER_PORT=8080
ENV LUMINOUS_SERVER_MODE=release

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
	CMD wget -qO- http://localhost:8080/healthz || exit 1

EXPOSE 8080

USER 65534

ENTRYPOINT ["./luminous"]
