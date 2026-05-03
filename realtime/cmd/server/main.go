package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/chrpnnkv/SmartBattle/backend-realtime/internal/auth"
	"github.com/chrpnnkv/SmartBattle/backend-realtime/internal/config"
	"github.com/chrpnnkv/SmartBattle/backend-realtime/internal/core"
	"github.com/chrpnnkv/SmartBattle/backend-realtime/internal/handler"
	"github.com/chrpnnkv/SmartBattle/backend-realtime/internal/room"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		logger.Error("ошибка загрузки конфигурации", "error", err)
		os.Exit(1)
	}

	if cfg.LogLevel == "debug" {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
		slog.SetDefault(logger)
	}

	logger.Info("конфигурация загружена",
		"addr", cfg.Addr(),
		"backend_core_url", cfg.BackendCoreURL,
		"max_participants", cfg.MaxParticipants,
	)

	authService := auth.NewService(cfg.JWTSecret)

	roomManager := room.NewManager(room.RoomConfig{
		RoomCodeLength:         cfg.RoomCodeLength,
		MaxParticipants:        cfg.MaxParticipants,
		DefaultQuestionTimeSec: cfg.DefaultQuestionTimeSec,
	}, logger)

	var coreClient *core.Client
	if cfg.BackendCoreURL != "" {
		coreClient = core.New(cfg.BackendCoreURL, cfg.BackendCoreInternalSecret, cfg.BackendCoreTimeout, logger)
	}

	mux := http.NewServeMux()
	h := handler.New(cfg, roomManager, authService, coreClient, logger)
	h.RegisterRoutes(mux)
	srv := &http.Server{
		Addr:         cfg.Addr(),
		Handler:      loggingMiddleware(logger, mux),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	go func() {
		logger.Info("сервер запущен", "addr", cfg.Addr())
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("ошибка сервера", "error", err)
			os.Exit(1)
		}
	}()

	<-shutdown
	logger.Info("получен сигнал завершения, останавливаем сервер...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("ошибка graceful shutdown", "error", err)
	}
	logger.Info("сервер остановлен")
}

// loggingMiddleware логирует каждый HTTP-запрос.
func loggingMiddleware(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)
		logger.Debug("http request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rw.status,
			"duration_ms", time.Since(start).Milliseconds(),
			"remote", r.RemoteAddr,
		)
	})
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

// Hijack пробрасывает интерфейс http.Hijacker до базового ResponseWriter.
// Без этого gorilla/websocket не может перехватить TCP-соединение
// и upgrade до WebSocket падает с "response does not implement http.Hijacker".
func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := rw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("underlying ResponseWriter does not implement http.Hijacker")
	}
	return h.Hijack()
}

// Flush пробрасывает http.Flusher (используется server-sent events / streaming).
func (rw *responseWriter) Flush() {
	if f, ok := rw.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}
