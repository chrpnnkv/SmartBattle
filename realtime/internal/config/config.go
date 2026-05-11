package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config — полная конфигурация сервиса.
type Config struct {
	// HTTP
	Host string
	Port string

	// JWT
	JWTSecret string

	// WebSocket
	WSReadLimit      int64
	WSWriteWait      time.Duration
	WSPongWait       time.Duration
	WSPingPeriod     time.Duration
	WSMaxMessageSize int64

	// Игровые комнаты
	RoomCodeLength  int
	MaxParticipants int

	// Rate limiting
	RateLimitMessages int
	RateLimitPeriod   time.Duration

	// Backend-core
	BackendCoreURL            string
	BackendCoreTimeout        time.Duration
	BackendCoreInternalSecret string

	// Таймер вопроса по умолчанию
	DefaultQuestionTimeSec int

	// Логирование
	LogLevel string
}

// Load загружает конфигурацию.
func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Host:                      getEnv("HOST", "0.0.0.0"),
		Port:                      getEnv("PORT", "8080"),
		JWTSecret:                 getEnv("JWT_SECRET", ""),
		WSReadLimit:               int64(getEnvInt("WS_READ_LIMIT", 4096)),
		WSWriteWait:               getEnvDuration("WS_WRITE_WAIT", 10*time.Second),
		WSPongWait:                getEnvDuration("WS_PONG_WAIT", 60*time.Second),
		WSMaxMessageSize:          int64(getEnvInt("WS_MAX_MESSAGE_SIZE", 4096)),
		RoomCodeLength:            getEnvInt("ROOM_CODE_LENGTH", 6),
		MaxParticipants:           getEnvInt("MAX_PARTICIPANTS", 100),
		RateLimitMessages:         getEnvInt("RATE_LIMIT_MESSAGES", 10),
		RateLimitPeriod:           getEnvDuration("RATE_LIMIT_PERIOD", time.Second),
		BackendCoreURL:            getEnv("BACKEND_CORE_URL", "http://localhost:8081"),
		BackendCoreTimeout:        getEnvDuration("BACKEND_CORE_TIMEOUT", 5*time.Second),
		BackendCoreInternalSecret: getEnv("BACKEND_CORE_INTERNAL_SECRET", getEnv("BACKEND_CORE_TOKEN", "")),
		DefaultQuestionTimeSec:    getEnvInt("DEFAULT_QUESTION_TIME_SEC", 30),
		LogLevel:                  getEnv("LOG_LEVEL", "info"),
	}

	cfg.WSPingPeriod = (cfg.WSPongWait * 9) / 10
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET не задан в переменных окружения")
	}

	return cfg, nil
}

// Addr возвращает адрес для прослушивания (host:port).
func (c *Config) Addr() string {
	return c.Host + ":" + c.Port
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return defaultVal
}

func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return defaultVal
}
