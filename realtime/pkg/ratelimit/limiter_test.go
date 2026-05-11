package ratelimit_test

import (
	"testing"
	"time"

	"github.com/chrpnnkv/SmartBattle/backend-realtime/pkg/ratelimit"
)

// TestLimiterAllowsWithinLimit проверяет, что сообщения в рамках лимита разрешены.
func TestLimiterAllowsWithinLimit(t *testing.T) {
	l := ratelimit.NewLimiter(5, time.Second)
	for i := 0; i < 5; i++ {
		if !l.Allow() {
			t.Errorf("сообщение %d должно быть разрешено", i+1)
		}
	}
}

// TestLimiterBlocksOverLimit проверяет, что превышение лимита блокируется.
func TestLimiterBlocksOverLimit(t *testing.T) {
	l := ratelimit.NewLimiter(3, time.Second)
	l.Allow()
	l.Allow()
	l.Allow()
	if l.Allow() {
		t.Error("четвёртое сообщение должно быть заблокировано")
	}
}

// TestLimiterRefillsOverTime проверяет пополнение токенов со временем.
func TestLimiterRefillsOverTime(t *testing.T) {
	l := ratelimit.NewLimiter(2, 100*time.Millisecond)
	l.Allow()
	l.Allow()
	if l.Allow() {
		t.Error("третий запрос должен быть заблокирован")
	}
	time.Sleep(150 * time.Millisecond)
	if !l.Allow() {
		t.Error("после ожидания запрос должен быть разрешён")
	}
}

// TestManagerAllowMultipleClients проверяет независимость лимитов для разных клиентов.
func TestManagerAllowMultipleClients(t *testing.T) {
	mgr := ratelimit.NewManager(2, time.Second)

	mgr.Allow("client1")
	mgr.Allow("client1")
	if !mgr.Allow("client2") {
		t.Error("client2 должен иметь независимый лимит")
	}
	if mgr.Allow("client1") {
		t.Error("client1 должен быть заблокирован")
	}
}

// TestManagerRemove проверяет удаление лимитера при отключении клиента.
func TestManagerRemove(t *testing.T) {
	mgr := ratelimit.NewManager(1, time.Second)
	mgr.Allow("client1")
	mgr.Remove("client1")
	if !mgr.Allow("client1") {
		t.Error("после Remove лимитер должен быть сброшен")
	}
}
