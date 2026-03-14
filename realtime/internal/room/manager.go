package room

import (
	"errors"
	"log/slog"
	"math/rand"
	"strings"
	"sync"
)

const roomCodeAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

// Manager — реестр игровых комнат.
type Manager struct {
	mu     sync.RWMutex
	rooms  map[string]*Room
	cfg    RoomConfig
	logger *slog.Logger
}

// NewManager создаёт новый менеджер комнат.
func NewManager(cfg RoomConfig, logger *slog.Logger) *Manager {
	return &Manager{
		rooms:  make(map[string]*Room),
		cfg:    cfg,
		logger: logger.With("component", "room_manager"),
	}
}

// Create создаёт новую игровую комнату с уникальным кодом.
func (m *Manager) Create(quizID, quizTitle string, questions []Question) (*Room, error) {
	if len(questions) == 0 {
		return nil, errors.New("квиз не содержит вопросов")
	}

	code := m.generateCode(6)

	room := New(code, quizID, quizTitle, questions, m.cfg, m.logger)

	m.mu.Lock()
	m.rooms[code] = room
	m.mu.Unlock()

	m.logger.Info("создана игровая комната",
		"code", code,
		"quiz_id", quizID,
		"questions", len(questions),
	)
	return room, nil
}

// Get возвращает комнату по коду. Второй аргумент false, если не найдена.
func (m *Manager) Get(code string) (*Room, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	r, ok := m.rooms[strings.ToUpper(code)]
	return r, ok
}

// Delete удаляет комнату из реестра.
func (m *Manager) Delete(code string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.rooms, code)
	m.logger.Info("комната удалена", "code", code)
}

// Count возвращает количество активных комнат.
func (m *Manager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.rooms)
}

// generateCode генерирует уникальный код комнаты заданной длины.
func (m *Manager) generateCode(length int) string {
	for {
		code := randomCode(length)
		m.mu.RLock()
		_, exists := m.rooms[code]
		m.mu.RUnlock()
		if !exists {
			return code
		}
	}
}

func randomCode(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = roomCodeAlphabet[rand.Intn(len(roomCodeAlphabet))]
	}
	return string(b)
}
