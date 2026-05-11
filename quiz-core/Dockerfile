FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN if [ ! -f admins.example.json ]; then \
    echo '{"admins": ["admin@hse.ru"]}' > admins.example.json; \
    fi

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /smartbattle-quiz-core ./cmd/api

FROM alpine:3.20

RUN apk add --no-cache ca-certificates && \
    addgroup -S app && adduser -S app -G app && \
    mkdir -p /uploads && chown app:app /uploads

COPY --from=builder /smartbattle-quiz-core /usr/local/bin/smartbattle-quiz-core

COPY --from=builder /app/migrations /migrations
ENV MIGRATIONS_DIR=/migrations

COPY --from=builder /app/admins.example.json /admins.json
ENV ADMINS_FILE=/admins.json
ENV UPLOADS_DIR=/uploads

USER app
EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/smartbattle-quiz-core"]
