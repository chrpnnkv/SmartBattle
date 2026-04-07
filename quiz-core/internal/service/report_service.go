package service

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"strconv"

	"github.com/chrpnnkv/SmartBattle/internal/models"
	"github.com/chrpnnkv/SmartBattle/internal/repository"
	"github.com/google/uuid"
)

type ReportService struct {
	repo *repository.ReportRepository
}

func NewReportService(repo *repository.ReportRepository) *ReportService {
	return &ReportService{repo: repo}
}

func (s *ReportService) SaveSessionReport(session *models.GameSession) error {
	return s.repo.Save(session)
}

func (s *ReportService) GetTeacherReports(hostID uuid.UUID) ([]models.GameSession, error) {
	return s.repo.GetByHostID(hostID)
}

func (s *ReportService) ExportCSV(id uuid.UUID) (*bytes.Buffer, error) {
	session, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	var reportData struct {
		Participants []struct {
			Nickname   string `json:"nickname"`
			TotalScore int    `json:"total_score"`
			Rank       int    `json:"rank"`
		} `json:"participants"`
	}

	if err := json.Unmarshal(session.ReportSnapshot, &reportData); err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	writer := csv.NewWriter(buf)

	_ = writer.Write([]string{"Rank", "Nickname", "Total Score"})
	for _, p := range reportData.Participants {
		_ = writer.Write([]string{
			strconv.Itoa(p.Rank),
			p.Nickname,
			strconv.Itoa(p.TotalScore),
		})
	}

	writer.Flush()
	return buf, nil
}
