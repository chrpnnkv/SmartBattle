package service

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"strings"
	"time"

	"github.com/chrpnnkv/SmartBattle/internal/models"
	"github.com/chrpnnkv/SmartBattle/internal/repository"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type ReportService struct {
	repo *repository.ReportRepository
}

type RealtimeParticipantResult struct {
	Name           string `json:"name"`
	Score          int    `json:"score"`
	CorrectAnswers int    `json:"correct_answers"`
	TotalQuestions int    `json:"total_questions"`
}

type RealtimeResultsPayload struct {
	QuizID     string                      `json:"quiz_id"`
	RoomCode   string                      `json:"room_code"`
	StartedAt  time.Time                   `json:"started_at"`
	FinishedAt time.Time                   `json:"finished_at"`
	Duration   int                         `json:"duration_sec"`
	Results    []RealtimeParticipantResult `json:"results"`
}

func NewReportService(repo *repository.ReportRepository) *ReportService {
	return &ReportService{repo: repo}
}

func (s *ReportService) SaveSessionReport(session *models.GameSession) error {
	return s.repo.SaveSessionReport(session)
}

func (s *ReportService) SaveRealtimeResults(payload *RealtimeResultsPayload) error {
	session, err := s.repo.GetByPIN(sanitizePIN(payload.RoomCode))
	if err != nil {
		return err
	}

	snapshot, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	session.Status = "finished"
	session.ReportSnapshot = datatypes.JSON(snapshot)
	if !payload.StartedAt.IsZero() {
		session.StartedAt = payload.StartedAt
	}
	if !payload.FinishedAt.IsZero() {
		finishedAt := payload.FinishedAt
		session.FinishedAt = &finishedAt
	}

	return s.repo.SaveSessionReport(session)
}

func (s *ReportService) GetTeacherReports(hostID uuid.UUID) ([]models.GameSession, error) {
	return s.repo.GetTeacherReports(hostID)
}

func (s *ReportService) GetReportByID(id uuid.UUID) (*models.GameSession, error) {
	return s.repo.GetByID(id)
}

func (s *ReportService) ExportCSV(id uuid.UUID) (*bytes.Buffer, error) {
	session, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	b := &bytes.Buffer{}
	w := csv.NewWriter(b)
	w.Write([]string{"ID", "QuizID", "Status", "PIN"})
	w.Write([]string{session.ID.String(), session.QuizID.String(), session.Status, session.PIN})
	w.Flush()

	return b, nil
}

func sanitizePIN(pin string) string {
	clean := strings.ReplaceAll(pin, " ", "")
	clean = strings.ReplaceAll(clean, "-", "")
	return strings.ToUpper(strings.ReplaceAll(clean, " ", ""))
}
