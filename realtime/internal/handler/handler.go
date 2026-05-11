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
			h.logger.Warn("WS join: невалидный токен", "client_id", c.ID, "error", err)
			writeErrorAndClose(conn, message.ErrCodeInvalidToken, "недействительный токен")
			return
		}
		userID = claims.UserID
		role = claims.Role
		if role == message.RoleTeacher && name == "" {
			// Сначала пробуем email из JWT, затем — fallback по UserID,
			// чтобы старые токены без email-claim не валили подключение.
			if claims.Email != "" {
				name = claims.Email
			} else {
				name = "Преподаватель"
			}
		}
	}

	if name == "" {
		h.logger.Warn("WS join: пустое имя", "client_id", c.ID, "role", role)
		writeErrorAndClose(conn, message.ErrCodeInvalidMessage, "поле name обязательно")
		return
	}
	if len(name) > 50 {
		h.logger.Warn("WS join: имя слишком длинное", "client_id", c.ID, "len", len(name))
		writeErrorAndClose(conn, message.ErrCodeInvalidMessage, "имя слишком длинное (макс. 50 символов)")
		return
	}

	roomCode := strings.ToUpper(strings.ReplaceAll(joinMsg.RoomCode, " ", ""))

	var rm *room.Room
	var found bool

	if role == message.RoleTeacher {
		rm, found = h.rooms.Get(roomCode)
		if !found {
			h.logger.Warn("WS join: комната не найдена (teacher)", "client_id", c.ID, "room_code", roomCode)
			writeErrorAndClose(conn, message.ErrCodeRoomNotFound, "комната не найдена: "+roomCode)
			return
		}
		if rm.IsFinished() {
			h.logger.Warn("WS join: сессия уже завершена (teacher)", "client_id", c.ID, "room_code", roomCode)
			writeErrorAndClose(conn, message.ErrCodeSessionNotActive, "сессия уже завершена")
			return
		}
	} else {
		rm, found = h.rooms.Get(roomCode)
		if !found {
			h.logger.Warn("WS join: комната не найдена (student)", "client_id", c.ID, "room_code", roomCode)
			writeErrorAndClose(conn, message.ErrCodeRoomNotFound, "комната не найдена: "+roomCode)
			return
		}
		if rm.IsFinished() {
			h.logger.Warn("WS join: сессия уже завершена", "client_id", c.ID, "room_code", roomCode)
			writeErrorAndClose(conn, message.ErrCodeSessionNotActive, "сессия уже завершена")
			return
		}
	}

	c.Name = name
	c.Role = role
	c.RoomCode = roomCode
	c.UserID = userID

	// Стабильный participantID (выданный backend-core при /api/sessions/join).
	// Если не передан — fallback на сгенерированный client.ID.
	stableParticipantID := strings.TrimSpace(joinMsg.ParticipantID)
	if stableParticipantID == "" {
		stableParticipantID = c.ID
	}

	if err := rm.AddClient(c, stableParticipantID); err != nil {
		// Различаем "комната заполнена" и "сессия уже завершена" — у них разные коды.
		errCode := message.ErrCodeRoomFull
		if rm.IsFinished() {
			errCode = message.ErrCodeSessionNotActive
		}
		h.logger.Warn("WS join: AddClient failed", "client_id", c.ID, "room_code", roomCode, "code", errCode, "error", err)
		writeErrorAndClose(conn, errCode, err.Error())
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

	// Снимок текущих участников — даём новому клиенту полную картину лобби,
	// чтобы счётчик «N студентов подключилось» был корректным сразу,
	// а не только после новых join-ов.
	participantsSnapshot := rm.GetParticipants()

	c.SendMsg(message.New(message.TypeJoined, message.JoinedPayload{
		RoomCode:       roomCode,
		Role:           role,
		Name:           name,
		QuizTitle:      rm.QuizTitle,
		TotalQuestions: len(rm.Questions),
		Participants:   participantsSnapshot,
		TotalCount:     len(participantsSnapshot),
	}))

	if role == message.RoleStudent {
		rm.BroadcastParticipantJoined(c, stableParticipantID)
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
	QuizMode  string            `json:"quiz_mode"`
	Questions []QuestionRequest `json:"questions"`
}

// OptionRequest — вариант ответа в запросе создания комнаты.
type OptionRequest struct {
	ID        string `json:"id"`
	Text      string `json:"text"`
	IsCorrect bool   `json:"is_correct"`
	Color     string `json:"color"`
}

// QuestionRequest — вопрос в запросе создания комнаты.
type QuestionRequest struct {
	ID           string          `json:"id"`
	Type         string          `json:"type,omitempty"`
	Text         string          `json:"text"`
	ImageURL     string          `json:"image_url,omitempty"`
	Score        int             `json:"score,omitempty"`
	Options      []OptionRequest `json:"options"`
	TimeLimitSec int             `json:"time_limit_sec"`
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
		for j, opt := range q.Options {
			id := opt.ID
			if id == "" {
				id = fmt.Sprintf("%s_o%d", q.ID, j)
			}
			color := opt.Color
			if color == "" {
				color = room.OptionColors[j%len(room.OptionColors)]
			}
			opts[j] = room.QuestionOption{
				ID:        id,
				Text:      opt.Text,
				IsCorrect: opt.IsCorrect,
				Color:     color,
			}
		}
		qType := q.Type
		if qType == "" {
			qType = "multiple_choice"
		}
		questions[i] = room.Question{
			ID:           q.ID,
			QuizID:       req.QuizID,
			Type:         qType,
			Text:         q.Text,
			ImageURL:     q.ImageURL,
			Score:        q.Score,
			Options:      opts,
			TimeLimitSec: q.TimeLimitSec,
			Order:        i,
		}
	}

	rm, err := h.rooms.Create(req.QuizID, req.QuizTitle, questions)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if req.QuizMode != "" {
		rm.Mode = req.QuizMode
	}
	// Запоминаем UserID учителя из JWT — нужно для последующего host_id в room info
	rm.TeacherUserID = claims.UserID

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
		"quiz_mode":  rm.Mode,
		"host_id":    rm.TeacherUserID,
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

// handleRoomByCode обрабатывает /api/rooms/{code} и /api/rooms/{code}/participants.
func (h *Handler) handleRoomByCode(w http.ResponseWriter, r *http.Request) {
	rest := strings.TrimPrefix(r.URL.Path, "/api/rooms/")
	if rest == "" {
		http.NotFound(w, r)
		return
	}

	parts := strings.SplitN(rest, "/", 2)
	code := strings.ToUpper(strings.ReplaceAll(parts[0], " ", ""))
	subpath := ""
	if len(parts) == 2 {
		subpath = parts[1]
	}

	rm, ok := h.rooms.Get(code)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "комната не найдена"})
		return
	}

	switch subpath {
	case "":
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"room_code":       rm.Code,
			"quiz_id":         rm.QuizID,
			"quiz_title":      rm.QuizTitle,
			"quiz_mode":       rm.Mode,
			"host_id":         rm.TeacherUserID,
			"status":          string(rm.GetStatus()),
			"participants":    rm.StudentCount(),
			"total_questions": len(rm.Questions),
		})
	case "participants":
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"participants":           rm.GetParticipants(),
			"current_question_index": rm.CurrentIndex(),
		})
	default:
		http.NotFound(w, r)
	}
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

	// Шлём корректный WS close frame, чтобы клиент получил wasClean=true
	// и не перезатёр серверную ошибку своим "[disconnected] Соединение прервано".
	closeFrame := websocket.FormatCloseMessage(websocket.CloseNormalClosure, msg)
	_ = conn.WriteControl(websocket.CloseMessage, closeFrame, time.Now().Add(time.Second))

	conn.Close()
}
