package room

import (
	"log/slog"
	"os"
	"testing"
	"time"
)

// calcScore — формула, которую видит студент в виде "Очки".
// Контракт:
//   - неверно → 0
//   - верно: 1000 базовых, минус скорость (быстрее = больше),
//     но не меньше 100 (минимум за правильный)
//   - верно за пределами лимита → 100 (поздний правильный)

func newTestRoomForScore() *Room {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	return &Room{
		cfg: RoomConfig{
			DefaultQuestionTimeSec: 30,
		},
		logger: logger,
	}
}

func TestCalcScore_IncorrectIsZero(t *testing.T) {
	r := newTestRoomForScore()
	got := r.calcScore(false, time.Now().Add(-2*time.Second), 30)
	if got != 0 {
		t.Errorf("expected 0 for incorrect answer, got %d", got)
	}
}

func TestCalcScore_InstantCorrectGetsHighScore(t *testing.T) {
	r := newTestRoomForScore()
	// Ответ "только что" — elapsed ≈ 0, ratio ≈ 0, score ≈ 1000.
	got := r.calcScore(true, time.Now(), 30)
	if got < 950 || got > 1000 {
		t.Errorf("expected score near 1000 for instant correct, got %d", got)
	}
}

func TestCalcScore_LateCorrectGetsFloor(t *testing.T) {
	r := newTestRoomForScore()
	// elapsed > limit → возвращается фиксированное минимальное значение 100.
	got := r.calcScore(true, time.Now().Add(-60*time.Second), 30)
	if got != 100 {
		t.Errorf("expected floor=100 for late-correct, got %d", got)
	}
}

func TestCalcScore_HalfwayCorrect(t *testing.T) {
	r := newTestRoomForScore()
	// elapsed ≈ 50% от лимита → score = 1000 - 900*0.5 = 550 (±небольшая дельта на время вызова).
	got := r.calcScore(true, time.Now().Add(-15*time.Second), 30)
	if got < 530 || got > 570 {
		t.Errorf("expected ~550 for halfway correct, got %d", got)
	}
}

func TestCalcScore_ZeroLimitFallsBackToDefault(t *testing.T) {
	r := newTestRoomForScore()
	// timeLimitSec=0 → используется DefaultQuestionTimeSec=30 из cfg.
	// Ответ за 15с при дефолтном лимите 30с → ~550.
	got := r.calcScore(true, time.Now().Add(-15*time.Second), 0)
	if got < 530 || got > 570 {
		t.Errorf("expected ~550 with default 30s limit, got %d", got)
	}
}

func TestCalcScore_NeverNegative(t *testing.T) {
	r := newTestRoomForScore()
	// Парадоксальные значения — "ответил в будущем". В формуле elapsed может стать отрицательным.
	// Score должен оставаться >= 100 (нижний пол).
	got := r.calcScore(true, time.Now().Add(60*time.Second), 30)
	if got < 100 {
		t.Errorf("expected score >= 100 even for weird timing, got %d", got)
	}
}
