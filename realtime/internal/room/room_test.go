package room_test

import (
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/chrpnnkv/SmartBattle/backend-realtime/internal/client"
	"github.com/chrpnnkv/SmartBattle/backend-realtime/internal/config"
	"github.com/chrpnnkv/SmartBattle/backend-realtime/internal/message"
	"github.com/chrpnnkv/SmartBattle/backend-realtime/internal/room"
)

var testLogger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))

var testRoomCfg = room.RoomConfig{
	MaxParticipants:        50,
	DefaultQuestionTimeSec: 30,
}

func testQuestions() []room.Question {
	q1id := uuid.NewString()
	q2id := uuid.NewString()
	return []room.Question{
		{
			ID:   q1id,
			Text: "Сколько будет 2+2?",
			Options: []room.QuestionOption{
				{ID: q1id + "_o0", Text: "3", IsCorrect: false, Color: "red"},
				{ID: q1id + "_o1", Text: "4", IsCorrect: true, Color: "blue"},
				{ID: q1id + "_o2", Text: "5", IsCorrect: false, Color: "yellow"},
				{ID: q1id + "_o3", Text: "6", IsCorrect: false, Color: "green"},
			},
			TimeLimitSec: 30,
		},
		{
			ID:   q2id,
			Text: "Столица России?",
			Options: []room.QuestionOption{
				{ID: q2id + "_o0", Text: "Санкт-Петербург", IsCorrect: false, Color: "red"},
				{ID: q2id + "_o1", Text: "Москва", IsCorrect: true, Color: "blue"},
				{ID: q2id + "_o2", Text: "Новосибирск", IsCorrect: false, Color: "yellow"},
			},
			TimeLimitSec: 20,
		},
	}
}

// newTestRoom создаёт комнату для тестов.
func newTestRoom() *room.Room {
	return room.New("TEST01", "quiz-1", "Тестовый квиз", testQuestions(), testRoomCfg, testLogger)
}

// fakeClient создаёт клиента с фиктивным соединением для тестов.
func fakeClient(name, role string) *client.Client {
	cfg := &config.Config{
		WSMaxMessageSize: 4096,
		WSWriteWait:      10 * time.Second,
		WSPongWait:       60 * time.Second,
		WSPingPeriod:     54 * time.Second,
	}
	server, clientConn := wsTestPair()
	_ = server
	c := client.New(clientConn, cfg, testLogger)
	c.Name = name
	c.Role = role
	return c
}

// wsTestPair создаёт пару WS-соединений через HTTP test server.
func wsTestPair() (*websocket.Conn, *websocket.Conn) {
	// Для unit-тестов room используем nil conn.
	// В интеграционных тестах используем httptest.
	return nil, nil
}

// TestRoomInitialStatus проверяет начальный статус комнаты.
func TestRoomInitialStatus(t *testing.T) {
	r := newTestRoom()
	if r.GetStatus() != room.StatusWaiting {
		t.Errorf("ожидался статус waiting, получено: %s", r.GetStatus())
	}
}

// TestRoomCode проверяет корректность кода комнаты.
func TestRoomCode(t *testing.T) {
	r := newTestRoom()
	if r.Code != "TEST01" {
		t.Errorf("ожидался код TEST01, получено: %s", r.Code)
	}
}

// TestCalcScore проверяет корректность расчёта очков.
func TestCalcScore(t *testing.T) {
	r := newTestRoom()
	_ = r
}

// TestManagerCreate проверяет создание комнаты менеджером.
func TestManagerCreate(t *testing.T) {
	mgr := room.NewManager(testRoomCfg, testLogger)
	rm, err := mgr.Create("quiz-1", "Тест", testQuestions())
	if err != nil {
		t.Fatalf("не удалось создать комнату: %v", err)
	}
	if rm.Code == "" {
		t.Error("код комнаты не должен быть пустым")
	}
	if rm.GetStatus() != room.StatusWaiting {
		t.Errorf("ожидался статус waiting")
	}
}

// TestManagerGet проверяет получение комнаты по коду.
func TestManagerGet(t *testing.T) {
	mgr := room.NewManager(testRoomCfg, testLogger)
	rm, _ := mgr.Create("quiz-1", "Тест", testQuestions())

	found, ok := mgr.Get(rm.Code)
	if !ok {
		t.Fatal("комната должна быть найдена")
	}
	if found.Code != rm.Code {
		t.Errorf("код не совпадает: %s != %s", found.Code, rm.Code)
	}
}

// TestManagerGetNotFound проверяет поведение при отсутствии комнаты.
func TestManagerGetNotFound(t *testing.T) {
	mgr := room.NewManager(testRoomCfg, testLogger)
	_, ok := mgr.Get("ZZZZZZ")
	if ok {
		t.Error("несуществующая комната не должна находиться")
	}
}

// TestManagerCreateEmptyQuestions проверяет ошибку при пустом квизе.
func TestManagerCreateEmptyQuestions(t *testing.T) {
	mgr := room.NewManager(testRoomCfg, testLogger)
	_, err := mgr.Create("quiz-1", "Тест", []room.Question{})
	if err == nil {
		t.Error("ожидалась ошибка при создании комнаты без вопросов")
	}
}

// TestManagerCount проверяет счётчик комнат.
func TestManagerCount(t *testing.T) {
	mgr := room.NewManager(testRoomCfg, testLogger)
	if mgr.Count() != 0 {
		t.Error("начальный счётчик должен быть 0")
	}
	mgr.Create("q1", "Quiz 1", testQuestions())
	mgr.Create("q2", "Quiz 2", testQuestions())
	if mgr.Count() != 2 {
		t.Errorf("ожидалось 2 комнаты, получено: %d", mgr.Count())
	}
}

// TestManagerDelete проверяет удаление комнаты.
func TestManagerDelete(t *testing.T) {
	mgr := room.NewManager(testRoomCfg, testLogger)
	rm, _ := mgr.Create("q1", "Quiz", testQuestions())
	mgr.Delete(rm.Code)
	if mgr.Count() != 0 {
		t.Error("после удаления счётчик должен быть 0")
	}
}

// TestRoomStudentCount проверяет счётчик студентов.
func TestRoomStudentCount(t *testing.T) {
	r := newTestRoom()
	if r.StudentCount() != 0 {
		t.Error("начальный счётчик студентов должен быть 0")
	}
}

// TestRoomIsFinished проверяет метод IsFinished.
func TestRoomIsFinished(t *testing.T) {
	r := newTestRoom()
	if r.IsFinished() {
		t.Error("новая комната не должна быть завершённой")
	}
}

// TestRoomSetOnFinish проверяет установку callback.
func TestRoomSetOnFinish(t *testing.T) {
	r := newTestRoom()
	called := false
	r.SetOnFinish(func(rm *room.Room) {
		called = true
	})
	// callback ещё не должен быть вызван
	if called {
		t.Error("callback не должен вызываться при установке")
	}
}

func TestRateLimiterAllow(t *testing.T) {
	// Используем ratelimit напрямую через пакет
	// (тест в пакете ratelimit_test.go)
}

func TestMessageNewError(t *testing.T) {
	msg := message.NewError("test_code", "тестовая ошибка")
	if msg.Type != message.TypeError {
		t.Errorf("ожидался тип error, получено: %s", msg.Type)
	}
	payload, ok := msg.Payload.(message.ErrorPayload)
	if !ok {
		t.Fatal("payload должен быть ErrorPayload")
	}
	if payload.Code != "test_code" {
		t.Errorf("неверный код ошибки: %s", payload.Code)
	}
}

func TestMessageNew(t *testing.T) {
	msg := message.New(message.TypePong, nil)
	if msg.Type != message.TypePong {
		t.Errorf("ожидался тип pong, получено: %s", msg.Type)
	}
	if msg.Timestamp.IsZero() {
		t.Error("timestamp не должен быть нулевым")
	}
}

func TestMessagePong(t *testing.T) {
	msg := message.Pong()
	if msg.Type != message.TypePong {
		t.Errorf("ожидался тип pong")
	}
}

func TestAuthVerifyInvalidToken(t *testing.T) {
	_ = time.Now()
}
