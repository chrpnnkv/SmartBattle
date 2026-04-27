package service

import (
	"bytes"
	"encoding/csv"

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
