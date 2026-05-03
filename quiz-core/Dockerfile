FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /smartbattle-quiz-core ./cmd/api

# Минимальная alpine — нужна для healthcheck (/bin/sh + busybox wget),
# scratch не подходит, потому что docker-compose использует CMD-SHELL.
FROM alpine:3.20

RUN apk add --no-cache ca-certificates && \
    addgroup -S app && adduser -S app -G app

COPY --from=builder /smartbattle-quiz-core /usr/local/bin/smartbattle-quiz-core

USER app
EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/smartbattle-quiz-core"]
