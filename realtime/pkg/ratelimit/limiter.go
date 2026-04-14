package ratelimit

import (
	"sync"
	"time"
)

// Limiter — per-client ограничитель частоты сообщений.
type Limiter struct {
	mu       sync.Mutex
	tokens   float64
	max      float64
	rate     float64
	lastTime time.Time
}

// NewLimiter создаёт ограничитель с заданным лимитом сообщений за period.
func NewLimiter(maxMessages int, period time.Duration) *Limiter {
	rate := float64(maxMessages) / period.Seconds()
	return &Limiter{
		tokens:   float64(maxMessages),
		max:      float64(maxMessages),
		rate:     rate,
		lastTime: time.Now(),
	}
}

// Allow проверяет, разрешено ли очередное сообщение.
func (l *Limiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(l.lastTime).Seconds()
	l.lastTime = now

	l.tokens += elapsed * l.rate
	if l.tokens > l.max {
		l.tokens = l.max
	}

	if l.tokens >= 1.0 {
		l.tokens--
		return true
	}
	return false
}

// Manager управляет ограничителями для множества клиентов.
type Manager struct {
	mu          sync.RWMutex
	limiters    map[string]*Limiter
	maxMessages int
	period      time.Duration
}

// NewManager создаёт менеджер rate limit.
func NewManager(maxMessages int, period time.Duration) *Manager {
	return &Manager{
		limiters:    make(map[string]*Limiter),
		maxMessages: maxMessages,
		period:      period,
	}
}

// Allow проверяет лимит для клиента clientID.
func (m *Manager) Allow(clientID string) bool {
	m.mu.RLock()
	l, ok := m.limiters[clientID]
	m.mu.RUnlock()

	if !ok {
		m.mu.Lock()
		l = NewLimiter(m.maxMessages, m.period)
		m.limiters[clientID] = l
		m.mu.Unlock()
	}

	return l.Allow()
}

// Remove удаляет ограничитель клиента при отключении.
func (m *Manager) Remove(clientID string) {
	m.mu.Lock()
	delete(m.limiters, clientID)
	m.mu.Unlock()
}
