package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	"github.com/chrpnnkv/SmartBattle/backend-realtime/internal/auth"
	"github.com/chrpnnkv/SmartBattle/backend-realtime/internal/client"
	"github.com/chrpnnkv/SmartBattle/backend-realtime/internal/config"
	"github.com/chrpnnkv/SmartBattle/backend-realtime/internal/core"
	"github.com/chrpnnkv/SmartBattle/backend-realtime/internal/message"
	"github.com/chrpnnkv/SmartBattle/backend-realtime/internal/room"
	"github.com/chrpnnkv/SmartBattle/backend-realtime/pkg/ratelimit"
)

// Handler — HTTP/WebSocket обработчик.
type Handler struct {
	cfg         *config.Config
	upgrader    websocket.Upgrader
	rooms       *room.Manager
	authService *auth.Service
	coreClient  *core.Client
	rateLimiter *ratelimit.Manager
	logger      *slog.Logger
}

// New создаёт новый Handler.
func New(
	cfg *config.Config,
	rooms *room.Manager,
	authService *auth.Service,
	coreClient *core.Client,
	logger *slog.Logger,
) *Handler {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
		HandshakeTimeout: 10 * time.Second,
	}

	return &Handler{
		cfg:         cfg,
		upgrader:    upgrader,
		rooms:       rooms,
		authService: authService,
		coreClient:  coreClient,
		rateLimiter: ratelimit.NewManager(cfg.RateLimitMessages, cfg.RateLimitPeriod),
		logger:      logger.With("component", "handler"),
	}
}

// RegisterRoutes регистрирует все маршруты в переданном мультиплексоре.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/ws", h.handleWebSocket)
	mux.HandleFunc("/ws/", h.handleWebSocket) // поддержка /ws/{sessionId} от фронтенда
	mux.HandleFunc("/api/rooms", h.handleRooms)
	mux.HandleFunc("/api/rooms/", h.handleRoomByCode)
	mux.HandleFunc("/health", h.handleHealth)
}

// handleWebSocket обрабатывает WebSocket-соединение.
func (h *Handler) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Warn("не удалось upgrade до WS", "error", err, "remote", r.RemoteAddr)
		return
	}

	c := client.New(conn, h.cfg, h.logger)
	h.logger.Info("новое WS-соединение", "client_id", c.ID, "remote", r.RemoteAddr)
	_ = conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	_, rawMsg, err := conn.ReadMessage()
	if err != nil {
		h.logger.Warn("таймаут handshake", "client_id", c.ID)
		conn.Close()
		return
	}
	_ = conn.SetReadDeadline(time.Time{})

	var joinMsg message.IncomingMessage
	if err := json.Unmarshal(rawMsg, &joinMsg); err != nil || joinMsg.Type != message.TypeJoin {
		writeErrorAndClose(conn, message.ErrCodeInvalidMessage, "первое сообщение должно быть join")
		return
	}

	role := message.RoleStudent
	name := strings.TrimSpace(joinMsg.Name)
	var userID string

	if joinMsg.Token != "" {
		claims, err := h.authService.Verify(joinMsg.Token)
		if err != nil {
			writeErrorAndClose(conn, message.ErrCodeInvalidToken, "недействительный токен")
			return
		}
		userID = claims.UserID
		role = claims.Role
		if role == message.RoleTeacher && name == "" {
			name = claims.Email
		}
	}

	if name == "" {
		writeErrorAndClose(conn, message.ErrCodeInvalidMessage, "поле name обязательно")
		return
	}
	if len(name) > 50 {
		writeErrorAndClose(conn, message.ErrCodeInvalidMessage, "имя слишком длинное (макс. 50 символов)")
		return
	}

	roomCode := strings.ToUpper(strings.TrimSpace(joinMsg.RoomCode))

	var rm *room.Room
	var found bool

	if role == message.RoleTeacher {
		rm, found = h.rooms.Get(roomCode)
		if !found {
			writeErrorAndClose(conn, message.ErrCodeRoomNotFound, "комната не найдена: "+roomCode)
			return
		}
	} else {
		rm, found = h.rooms.Get(roomCode)
		if !found {
			writeErrorAndClose(conn, message.ErrCodeRoomNotFound, "комната не найдена: "+roomCode)
			return
		}
		if rm.IsFinished() {
			writeErrorAndClose(conn, message.ErrCodeSessionNotActive, "сессия уже завершена")
			return
		}
	}

	c.Name = name
	c.Role = role
	c.RoomCode = roomCode
	c.UserID = userID

	if err := rm.AddClient(c, c.ID); err != nil {
		writeErrorAndClose(conn, message.ErrCodeRoomFull, err.Error())
		return
	}

	c.OnMessage = func(cl *client.Client, msg message.IncomingMessage) {
		// Rate limiting
		if !h.rateLimiter.Allow(cl.ID) {
			cl.SendMsg(message.NewError(message.ErrCodeRateLimitExceeded, "слишком много сообщений"))
			return
		}
		rm.HandleMessage(cl, msg)
	}

	c.OnDisconnect = func(cl *client.Client) {
		h.rateLimiter.Remove(cl.ID)
		rm.RemoveClient(cl.ID)
		h.logger.Info("клиент отключился", "client_id", cl.ID, "name", cl.Name)
	}

	c.SendMsg(message.New(message.TypeJoined, message.JoinedPayload{
		RoomCode:       roomCode,
		Role:           role,
		Name:           name,
		QuizTitle:      rm.QuizTitle,
		TotalQuestions: len(rm.Questions),
	}))

	if role == message.RoleStudent {
		rm.BroadcastParticipantJoined(c, c.ID)
	}
	c.Start()

	h.logger.Info("клиент вошёл в комнату",
		"client_id", c.ID,
		"name", name,
		"role", role,
		"room_code", roomCode,
	)
}

// CreateRoomRequest — тело запроса на создание комнаты.
type CreateRoomRequest struct {
	QuizID    string            `json:"quiz_id"`
	QuizTitle string            `json:"quiz_title"`
	Questions []QuestionRequest `json:"questions"`
}

// QuestionRequest — вопрос в запросе создания комнаты.
type QuestionRequest struct {
	ID           string   `json:"id"`
	Text         string   `json:"text"`
	Options      []string `json:"options"`
	CorrectIndex int      `json:"correct_index"`
	TimeLimitSec int      `json:"time_limit_sec"`
}

// handleRooms обрабатывает /api/rooms.
func (h *Handler) handleRooms(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.createRoom(w, r)
	case http.MethodGet:
		h.listRooms(w, r)
	default:
		http.Error(w, "метод не поддерживается", http.StatusMethodNotAllowed)
	}
}

// createRoom создаёт новую игровую комнату.
func (h *Handler) createRoom(w http.ResponseWriter, r *http.Request) {
	claims, err := h.extractClaims(r)
	if err != nil || !claims.IsTeacher() {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "требуется авторизация преподавателя"})
		return
	}

	var req CreateRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "невалидный JSON"})
		return
	}

	if len(req.Questions) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "список вопросов пуст"})
		return
	}

	questions := make([]room.Question, len(req.Questions))
	for i, q := range req.Questions {
		opts := make([]room.QuestionOption, len(q.Options))
		for j, text := range q.Options {
			opts[j] = room.QuestionOption{
				ID:        fmt.Sprintf("%s_o%d", q.ID, j),
				Text:      text,
				IsCorrect: j == q.CorrectIndex,
				Color:     room.OptionColors[j%len(room.OptionColors)],
			}
		}
		questions[i] = room.Question{
			ID:           q.ID,
			Text:         q.Text,
			Options:      opts,
			TimeLimitSec: q.TimeLimitSec,
		}
	}

	rm, err := h.rooms.Create(req.QuizID, req.QuizTitle, questions)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	rm.SetOnFinish(func(r *room.Room) {
		if h.coreClient != nil {
			if err := h.coreClient.SaveResults(r); err != nil {
				h.logger.Error("ошибка передачи итогов в backend-core",
					"room_code", r.Code,
					"error", err,
				)
			}
		}
		go func() {
			time.Sleep(5 * time.Minute)
			h.rooms.Delete(r.Code)
		}()
	})

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"room_code":  rm.Code,
		"quiz_id":    rm.QuizID,
		"quiz_title": rm.QuizTitle,
		"status":     string(rm.GetStatus()),
		"ws_url":     "/ws",
	})
}

// listRooms возвращает количество активных комнат (для мониторинга).
func (h *Handler) listRooms(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"active_rooms": h.rooms.Count(),
	})
}

// handleRoomByCode обрабатывает /api/rooms/{code}.
func (h *Handler) handleRoomByCode(w http.ResponseWriter, r *http.Request) {
	code := strings.TrimPrefix(r.URL.Path, "/api/rooms/")
	code = strings.ToUpper(strings.TrimSpace(code))

	if code == "" {
		http.NotFound(w, r)
		return
	}

	rm, ok := h.rooms.Get(code)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "комната не найдена"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"room_code":       rm.Code,
		"quiz_title":      rm.QuizTitle,
		"status":          string(rm.GetStatus()),
		"participants":    rm.StudentCount(),
		"total_questions": len(rm.Questions),
	})
}

// handleHealth возвращает статус сервиса.
func (h *Handler) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":       "ok",
		"active_rooms": h.rooms.Count(),
		"timestamp":    time.Now().UTC(),
	})
}

func (h *Handler) extractClaims(r *http.Request) (*auth.Claims, error) {
	authHeader := r.Header.Get("Authorization")
	token := strings.TrimPrefix(authHeader, "Bearer ")
	return h.authService.Verify(token)
}

func writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErrorAndClose(conn *websocket.Conn, code, msg string) {
	errMsg := message.NewError(code, msg)
	data, _ := json.Marshal(errMsg)
	_ = conn.WriteMessage(websocket.TextMessage, data)
	conn.Close()
}
