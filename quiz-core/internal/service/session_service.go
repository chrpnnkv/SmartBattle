package service

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/chrpnnkv/SmartBattle/internal/models"
	"github.com/chrpnnkv/SmartBattle/internal/realtime"
	"github.com/chrpnnkv/SmartBattle/internal/repository"
	"github.com/google/uuid"
)

type SessionService struct {
	repo     *repository.SessionRepository
	quizRepo *repository.QuizRepository
	rt       realtime.Client
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

func NewSessionService(
	repo *repository.SessionRepository,
	quizRepo *repository.QuizRepository,
	rt realtime.Client,
) *SessionService {
	return &SessionService{repo: repo, quizRepo: quizRepo, rt: rt}
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

	rtReq := buildCreateRoomRequest(quiz, mode)

	rtResp, err := s.rt.CreateRoom(context.Background(), hostID, rtReq)
	if err != nil {
		return nil, err
	}

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

	log.Printf("CreateSession: id=%s host_id=%s pin=%s mode=%s",
		session.ID, session.HostID, session.PIN, session.Mode)
	return session, nil
}

func (s *SessionService) JoinSession(pin string) (*JoinSessionResult, error) {
	cleanPIN := sanitizePIN(pin)

	session, err := s.repo.GetByPIN(cleanPIN)
	if err != nil {
		if !errors.Is(err, repository.ErrNotFound) {
			return nil, err
		}

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
	info, err := s.rt.GetRoom(context.Background(), pin)
	if err != nil {
		return nil, repository.ErrNotFound
	}

	quizID, err := uuid.Parse(info.QuizID)
	if err != nil {
		return nil, err
	}

	status := "waiting"
	if info.Status == "active" || info.Status == "finished" {
		status = info.Status
	}

	hostID := uuid.Nil
	if info.HostID != "" {
		if parsed, parseErr := uuid.Parse(info.HostID); parseErr == nil {
			hostID = parsed
		}
	}
	if hostID == uuid.Nil {
		log.Printf("WARN restoreSessionFromRealtime: pin=%s recovered with host_id=Nil — отчёт не привяжется к учителю", pin)
	}

	restored := &models.GameSession{
		QuizID: quizID,
		HostID: hostID,
		PIN:    pin,
		Status: status,
	}

	if err := s.repo.Create(restored); err != nil {
		if existing, getErr := s.repo.GetByPIN(pin); getErr == nil {
			return existing, nil
		}
		return nil, err
	}
	return restored, nil
}

func (s *SessionService) FetchLiveRoom(pin string) (realtime.RoomParticipants, bool) {
	body, err := s.rt.GetParticipants(context.Background(), pin)
	if err != nil {
		return realtime.RoomParticipants{}, false
	}
	return body, true
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

func buildCreateRoomRequest(quiz *models.Quiz, mode string) realtime.CreateRoomRequest {
	req := realtime.CreateRoomRequest{
		QuizID:    quiz.ID.String(),
		QuizTitle: quiz.Title,
		QuizMode:  mode,
		Questions: make([]realtime.Question, 0, len(quiz.Questions)),
	}
	for _, q := range quiz.Questions {
		rtQ := realtime.Question{
			ID:           q.ID.String(),
			Type:         q.Type,
			Text:         q.Text,
			ImageURL:     q.ImageURL,
			Score:        q.Score,
			TimeLimitSec: q.TimerSec,
			Options:      make([]realtime.Option, 0, len(q.Options)),
		}
		for _, opt := range q.Options {
			rtQ.Options = append(rtQ.Options, realtime.Option{
				ID:        opt.ID.String(),
				Text:      opt.Text,
				IsCorrect: opt.IsCorrect,
				Color:     opt.Color,
			})
		}
		req.Questions = append(req.Questions, rtQ)
	}
	return req
}
