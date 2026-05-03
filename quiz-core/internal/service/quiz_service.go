package service

import (
	"github.com/chrpnnkv/SmartBattle/internal/models"
	"github.com/chrpnnkv/SmartBattle/internal/repository"
	"github.com/google/uuid"
)

type QuizService struct {
	repo *repository.QuizRepository
}

func NewQuizService(repo *repository.QuizRepository) *QuizService {
	return &QuizService{repo: repo}
}

func (s *QuizService) CreateQuiz(quiz *models.Quiz) error {
	return s.repo.Create(quiz)
}

func (s *QuizService) GetQuizByID(id uuid.UUID) (*models.Quiz, error) {
	return s.repo.GetByID(id)
}

func (s *QuizService) GetTeacherQuizzes(teacherID uuid.UUID) ([]models.Quiz, error) {
	return s.repo.GetByTeacherID(teacherID)
}

func (s *QuizService) GetPublicQuizzes() ([]models.Quiz, error) {
	return s.repo.GetPublic()
}

func (s *QuizService) UpdateQuiz(quiz *models.Quiz) error {
	return s.repo.Update(quiz)
}

func (s *QuizService) DeleteQuiz(id uuid.UUID) error {
	return s.repo.Delete(id)
}
