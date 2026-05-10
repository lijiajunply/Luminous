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
COPY --from=builder /build/data ./data

ENV LUMINOUS_SERVER_PORT=8080
ENV LUMINOUS_SERVER_MODE=release

EXPOSE 8080

ENTRYPOINT ["./luminous"]
