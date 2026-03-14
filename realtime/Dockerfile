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
FROM scratch

# Копируем корневые сертификаты для HTTPS-запросов к backend-core
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Копируем бинарник
COPY --from=builder /smartbattle-realtime /smartbattle-realtime

EXPOSE 8080

ENTRYPOINT ["/smartbattle-realtime"]
