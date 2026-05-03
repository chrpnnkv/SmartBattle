package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/chrpnnkv/SmartBattle/internal/config"
	"github.com/chrpnnkv/SmartBattle/internal/models"
	"github.com/chrpnnkv/SmartBattle/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SessionService struct {
	repo     *repository.SessionRepository
	quizRepo *repository.QuizRepository
	cfg      *config.Config
}

func NewSessionService(repo *repository.SessionRepository, quizRepo *repository.QuizRepository, cfg *config.Config) *SessionService {
	return &SessionService{repo: repo, quizRepo: quizRepo, cfg: cfg}
}

type realtimeOption struct {
	ID        string `json:"id"`
	Text      string `json:"text"`
	IsCorrect bool   `json:"is_correct"`
	Color     string `json:"color"`
}

type realtimeQuestion struct {
	ID           string           `json:"id"`
	Type         string           `json:"type"`
	Text         string           `json:"text"`
	ImageURL     string           `json:"image_url,omitempty"`
	Score        int              `json:"score"`
	Options      []realtimeOption `json:"options"`
	TimeLimitSec int              `json:"time_limit_sec"`
}

type realtimeReq struct {
	QuizID    string             `json:"quiz_id"`
	QuizTitle string             `json:"quiz_title"`
	QuizMode  string             `json:"quiz_mode,omitempty"`
	Questions []realtimeQuestion `json:"questions"`
}

type realtimeResp struct {
	RoomCode string `json:"room_code"`
	Error    string `json:"error"`
}

type JoinSessionQuizInfo struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Mode  string `json:"mode"`
}

type JoinSessionResult struct {
	SessionID     string              `json:"sessionId"`
	ParticipantID string              `json:"participantId"`
	Status        string              `json:"status"`
	Quiz          JoinSessionQuizInfo `json:"quiz"`
}

type realtimeRoomInfoResponse struct {
	RoomCode     string `json:"room_code"`
	QuizID       string `json:"quiz_id"`
	QuizMode     string `json:"quiz_mode"`
	HostID       string `json:"host_id"`
	Status       string `json:"status"`
	Error        string `json:"error"`
	Participants int    `json:"participants"`
}

// SessionDTO — обогащённое представление сессии для фронтенда.
// Поля Mode/CurrentQuestionIndex/TotalQuestions/Participants собираются
// из БД + realtime-комнаты, чтобы фронтенд получал то, что описано в TS-типе GameSession.
type SessionDTO struct {
	ID                   string           `json:"id"`
	QuizID               string           `json:"quizId"`
	HostID               string           `json:"hostId"`
	PIN                  string           `json:"pin"`
	Status               string           `json:"status"`
	Mode                 string           `json:"mode"`
	StartedAt            time.Time        `json:"startedAt"`
	FinishedAt           *time.Time       `json:"finishedAt"`
	CurrentQuestionIndex int              `json:"currentQuestionIndex"`
	TotalQuestions       int              `json:"totalQuestions"`
	Participants         []SessionDTOPart `json:"participants"`
}

type SessionDTOPart struct {
	ID             string `json:"id"`
	Nickname       string `json:"nickname"`
	AvatarInitials string `json:"avatarInitials"`
	AvatarColor    string `json:"avatarColor"`
	Score          int    `json:"score"`
	AnsweredCount  int    `json:"answeredCount"`
}

// BuildSessionDTO собирает SessionDTO для фронтенда.
// Если сессия активна — пытается дёрнуть realtime, чтобы добрать participants/CurrentQuestionIndex.
// Если realtime недоступен — отдаёт только данные из БД (participants пустой).
func (s *SessionService) BuildSessionDTO(session *models.GameSession) SessionDTO {
	dto := SessionDTO{
		ID:           session.ID.String(),
		QuizID:       session.QuizID.String(),
		HostID:       session.HostID.String(),
		PIN:          session.PIN,
		Status:       session.Status,
		Mode:         session.Mode,
		StartedAt:    session.StartedAt,
		FinishedAt:   session.FinishedAt,
		Participants: []SessionDTOPart{},
	}

	// totalQuestions — из квиза.
	if quiz, err := s.quizRepo.GetByID(session.QuizID); err == nil {
		dto.TotalQuestions = len(quiz.Questions)
	}

	// participants/currentQuestionIndex — пытаемся дёрнуть из realtime.
	if session.Status == "waiting" || session.Status == "active" {
		if parts, idx, ok := s.fetchRealtimeRoomDetails(session.PIN); ok {
			dto.Participants = parts
			dto.CurrentQuestionIndex = idx
		}
	}

	return dto
}

// fetchRealtimeRoomDetails — best-effort запрос участников и текущего вопроса из realtime.
func (s *SessionService) fetchRealtimeRoomDetails(pin string) ([]SessionDTOPart, int, bool) {
	req, err := http.NewRequest(http.MethodGet, s.cfg.RealtimeURL+"/api/rooms/"+pin+"/participants", nil)
	if err != nil {
		return nil, 0, false
	}
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, 0, false
	}
	var body struct {
		Participants         []SessionDTOPart `json:"participants"`
		CurrentQuestionIndex int              `json:"current_question_index"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, 0, false
	}
	if body.Participants == nil {
		body.Participants = []SessionDTOPart{}
	}
	return body.Participants, body.CurrentQuestionIndex, true
}

func (s *SessionService) CreateSession(quizID, hostID uuid.UUID, mode string) (*models.GameSession, error) {
	quiz, err := s.quizRepo.GetByID(quizID)
	if err != nil {
		return nil, errors.New("quiz not found")
	}

	if mode == "" {
		mode = quiz.Mode
	}
	if mode == "" {
		mode = "teacher_paced"
	}

	reqPayload := realtimeReq{
		QuizID:    quiz.ID.String(),
		QuizTitle: quiz.Title,
		QuizMode:  mode,
		Questions: make([]realtimeQuestion, 0, len(quiz.Questions)),
	}

	for _, q := range quiz.Questions {
		rtQ := realtimeQuestion{
			ID:           q.ID.String(),
			Type:         q.Type,
			Text:         q.Text,
			ImageURL:     q.ImageURL,
			Score:        q.Score,
			TimeLimitSec: q.TimerSec,
			Options:      make([]realtimeOption, 0, len(q.Options)),
		}
		for _, opt := range q.Options {
			rtQ.Options = append(rtQ.Options, realtimeOption{
				ID:        opt.ID.String(),
				Text:      opt.Text,
				IsCorrect: opt.IsCorrect,
				Color:     opt.Color,
			})
		}
		reqPayload.Questions = append(reqPayload.Questions, rtQ)
	}

	body, _ := json.Marshal(reqPayload)

	// Service-to-service JWT для realtime. email опционален и оставлен пустым,
	// он используется realtime только как fallback-имя для учителя.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": hostID.String(),
		"role":    "teacher",
		"exp":     time.Now().Add(time.Minute * 5).Unix(),
	})
	tokenString, err := token.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil || tokenString == "" {
		return nil, fmt.Errorf("failed to sign service token: %w", err)
	}

	req, _ := http.NewRequest(http.MethodPost, s.cfg.RealtimeURL+"/api/rooms", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tokenString)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to contact realtime server: %v", err)
	}
	defer resp.Body.Close()

	var rtResp realtimeResp
	if err := json.NewDecoder(resp.Body).Decode(&rtResp); err != nil {
		return nil, errors.New("invalid response from realtime server")
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("realtime server error: %s", rtResp.Error)
	}

	// Сохраняем АБСОЛЮТНО ЧИСТЫЙ ПИН в БД
	cleanRoomCode := sanitizePIN(rtResp.RoomCode)

	session := &models.GameSession{
		QuizID: quizID,
		HostID: hostID,
		PIN:    cleanRoomCode,
		Status: "waiting",
		Mode:   mode,
	}

	if err := s.repo.Create(session); err != nil {
		return nil, err
	}

	fmt.Printf("CreateSession: id=%s host_id=%s pin=%s mode=%s\n",
		session.ID, session.HostID, session.PIN, session.Mode)
	return session, nil
}

func (s *SessionService) JoinSession(pin string) (*JoinSessionResult, error) {
	cleanPIN := sanitizePIN(pin)

	session, err := s.repo.GetByPIN(cleanPIN)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}

		// Self-healing: если БД временно "потеряла" сессию, подтягиваем комнату из realtime
		session, err = s.restoreSessionFromRealtime(cleanPIN)
		if err != nil {
			return nil, errors.New("Сессия с таким PIN не найдена")
		}
	}
	if session.Status == "finished" {
		return nil, errors.New("Сессия уже завершена")
	}

	quiz, err := s.quizRepo.GetByID(session.QuizID)
	if err != nil {
		return nil, errors.New("quiz not found")
	}

	participantID := uuid.New().String()
	return &JoinSessionResult{
		SessionID:     session.ID.String(),
		ParticipantID: participantID,
		Status:        session.Status,
		Quiz: JoinSessionQuizInfo{
			ID:    quiz.ID.String(),
			Title: quiz.Title,
			Mode:  quiz.Mode,
		},
	}, nil
}

func (s *SessionService) restoreSessionFromRealtime(pin string) (*models.GameSession, error) {
	req, err := http.NewRequest(http.MethodGet, s.cfg.RealtimeURL+"/api/rooms/"+pin, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, gorm.ErrRecordNotFound
	}

	var rt realtimeRoomInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&rt); err != nil {
		return nil, err
	}

	quizID, err := uuid.Parse(rt.QuizID)
	if err != nil {
		return nil, err
	}

	status := "waiting"
	if rt.Status == "active" || rt.Status == "finished" {
		status = rt.Status
	}

	hostID := uuid.Nil
	if rt.HostID != "" {
		if parsed, parseErr := uuid.Parse(rt.HostID); parseErr == nil {
			hostID = parsed
		}
	}

	restored := &models.GameSession{
		QuizID: quizID,
		HostID: hostID,
		PIN:    pin,
		Status: status,
	}

	if err := s.repo.Create(restored); err != nil {
		// если конкурентный запрос уже создал запись — просто читаем ее
		existing, getErr := s.repo.GetByPIN(pin)
		if getErr == nil {
			return existing, nil
		}
		return nil, err
	}

	return restored, nil
}

func (s *SessionService) GetSession(id uuid.UUID) (*models.GameSession, error) {
	return s.repo.GetByID(id)
}

func (s *SessionService) ChangeStatus(id uuid.UUID, newStatus string) error {
	session, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	session.Status = newStatus
	if newStatus == "finished" {
		now := time.Now()
		session.FinishedAt = &now
	}
	return s.repo.Update(session)
}
