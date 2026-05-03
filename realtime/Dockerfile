# ── Этап сборки ────────────────────────────────────────────
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Копируем go.mod и go.sum для кэширования зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем статический бинарник
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /smartbattle-realtime ./cmd/server

# ── Финальный образ ─────────────────────────────────────────
# Минимальная alpine вместо scratch: нужен /bin/sh и wget для healthcheck.
FROM alpine:3.20

RUN apk add --no-cache ca-certificates && \
    addgroup -S app && adduser -S app -G app

COPY --from=builder /smartbattle-realtime /usr/local/bin/smartbattle-realtime

USER app
EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/smartbattle-realtime"]