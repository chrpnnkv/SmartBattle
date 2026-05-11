package service

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"sort"
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

func (s *ReportService) GetAllReports() ([]models.GameSession, error) {
	return s.repo.GetAllReports()
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
	defer w.Flush()

	_ = w.Write([]string{"Session ID", session.ID.String()})
	_ = w.Write([]string{"Quiz ID", session.QuizID.String()})
	_ = w.Write([]string{"PIN", session.PIN})
	_ = w.Write([]string{"Status", session.Status})
	if !session.StartedAt.IsZero() {
		_ = w.Write([]string{"Started at", session.StartedAt.UTC().Format(time.RFC3339)})
	}
	if session.FinishedAt != nil && !session.FinishedAt.IsZero() {
		_ = w.Write([]string{"Finished at", session.FinishedAt.UTC().Format(time.RFC3339)})
	}
	_ = w.Write([]string{})

	_ = w.Write([]string{"Rank", "Nickname", "Score", "Correct", "Total", "Accuracy %"})

	if len(session.ReportSnapshot) > 0 {
		var snap RealtimeResultsPayload
		if err := json.Unmarshal(session.ReportSnapshot, &snap); err == nil {
			results := make([]RealtimeParticipantResult, len(snap.Results))
			copy(results, snap.Results)
			sort.SliceStable(results, func(i, j int) bool {
				return results[i].Score > results[j].Score
			})

			for i, r := range results {
				accuracy := 0
				if r.TotalQuestions > 0 {
					accuracy = int(float64(r.CorrectAnswers) * 100.0 / float64(r.TotalQuestions))
				}
				_ = w.Write([]string{
					fmt.Sprintf("%d", i+1),
					r.Name,
					fmt.Sprintf("%d", r.Score),
					fmt.Sprintf("%d", r.CorrectAnswers),
					fmt.Sprintf("%d", r.TotalQuestions),
					fmt.Sprintf("%d", accuracy),
				})
			}
		}
	}

	return b, nil
}

func sanitizePIN(pin string) string {
	clean := strings.ReplaceAll(pin, " ", "")
	clean = strings.ReplaceAll(clean, "-", "")
	return strings.ToUpper(strings.ReplaceAll(clean, " ", ""))
}
