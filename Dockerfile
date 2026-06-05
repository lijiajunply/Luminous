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

COPY --from=builder /build/luminous .

# ── Configuration via environment variables (LUMINOUS_ prefix) ──
# Server:
#   LUMINOUS_SERVER_PORT=8080
#   LUMINOUS_SERVER_MODE=release
#   LUMINOUS_SERVER_CORS_ORIGIN=https://myapp.example.com
# Auth:
#   LUMINOUS_AUTH_ADMIN_TOKEN=<your-secret-token>
# Database (DSN takes priority):
#   LUMINOUS_DATABASE_DSN=postgresql://user:password@host:port/dbname
#   LUMINOUS_DATABASE_HOST / PORT / USER / PASSWORD / DBNAME / SSLMODE
#   LUMINOUS_DATABASE_POOL_MAX_CONNS=20
#   LUMINOUS_DATABASE_POOL_MIN_CONNS=5
# Release proxy:
#   LUMINOUS_RELEASE_API_URL=<full-override-url>
#   LUMINOUS_RELEASE_APP_UUID=<app-uuid>
#   LUMINOUS_RELEASE_CHANNEL_ID=<channel-id>

ENV LUMINOUS_SERVER_PORT=8080
ENV LUMINOUS_SERVER_MODE=release

EXPOSE 8080

ENTRYPOINT ["./luminous"]
