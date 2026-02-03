package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"quiz-core/internal/config"
	"quiz-core/internal/models"
	"quiz-core/pkg/auth"
	"time"

	"github.com/google/uuid"
)

type QuizRepository interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	UpdatePassword(ctx context.Context, userID uuid.UUID, hash string) error
	SetResetToken(ctx context.Context, userID uuid.UUID, token string, expiry time.Time) error
	GetUserByResetToken(ctx context.Context, token string) (*models.User, error)
	ResetPassword(ctx context.Context, userID uuid.UUID, newHash string) error

	CreateQuiz(ctx context.Context, quiz *models.Quiz) error
	UpdateQuiz(ctx context.Context, quiz *models.Quiz) error
	GetQuizzesByTeacher(ctx context.Context, teacherID uuid.UUID, limit, offset int) ([]models.Quiz, error)
	GetPublicQuizzes(ctx context.Context, limit, offset int) ([]models.Quiz, error)
	GetFullQuiz(ctx context.Context, id int64) (*models.Quiz, error)
	DeleteQuiz(ctx context.Context, id int64, teacherID uuid.UUID) error

	SaveGameSession(ctx context.Context, session *models.GameSession) error
	GetSessionsByTeacher(ctx context.Context, teacherID uuid.UUID) ([]models.GameSession, error)
	GetSessionByID(ctx context.Context, id uuid.UUID) (*models.GameSession, error)
}

type Service struct {
	repo QuizRepository
	cfg  *config.Config
}

func NewService(repo QuizRepository, cfg *config.Config) *Service {
	return &Service{repo: repo, cfg: cfg}
}

func (s *Service) RegisterUser(ctx context.Context, email, password string) error {
	hash, err := auth.HashPassword(password)
	if err != nil {
		return err
	}
	user := &models.User{
		Email:        email,
		PasswordHash: hash,
	}
	return s.repo.CreateUser(ctx, user)
}

func (s *Service) LoginUser(ctx context.Context, email, password string) (string, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", errors.New("invalid credentials")
	}
	if !auth.CheckPassword(password, user.PasswordHash) {
		return "", errors.New("invalid credentials")
	}
	return auth.GenerateToken(user.ID, user.Role, s.cfg.JWTSecret)
}

func (s *Service) GetUserProfile(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	return s.repo.GetUserByID(ctx, userID)
}

func (s *Service) ChangePassword(ctx context.Context, userID uuid.UUID, oldPass, newPass string) error {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	if !auth.CheckPassword(oldPass, user.PasswordHash) {
		return errors.New("wrong old password")
	}
	newHash, err := auth.HashPassword(newPass)
	if err != nil {
		return err
	}
	return s.repo.UpdatePassword(ctx, userID, newHash)
}

func (s *Service) ForgotPassword(ctx context.Context, email string) error {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		log.Printf("ForgotPassword: user not found for %s", email)
		return nil
	}
	token := uuid.New().String()
	expiry := time.Now().Add(15 * time.Minute)
	if err := s.repo.SetResetToken(ctx, user.ID, token, expiry); err != nil {
		return err
	}
	log.Printf("=== PASSWORD RESET EMAIL ===")
	log.Printf("To: %s", email)
	log.Printf("Token: %s", token)
	log.Printf("============================")
	return nil
}

func (s *Service) ResetPassword(ctx context.Context, token, newPassword string) error {
	user, err := s.repo.GetUserByResetToken(ctx, token)
	if err != nil {
		return errors.New("invalid or expired token")
	}
	newHash, err := auth.HashPassword(newPassword)
	if err != nil {
		return err
	}
	return s.repo.ResetPassword(ctx, user.ID, newHash)
}

func (s *Service) validateQuiz(quiz *models.Quiz) error {
	if quiz.Title == "" {
		return errors.New("quiz title is required")
	}
	if len(quiz.Questions) == 0 {
		return errors.New("quiz must have at least one question")
	}
	for i, q := range quiz.Questions {
		if q.Text == "" {
			return fmt.Errorf("question %d must have text", i+1)
		}
		if q.Type == "single_choice" || q.Type == "multiple_choice" {
			if len(q.Options) < 2 {
				return fmt.Errorf("question %d must have at least 2 options", i+1)
			}
			hasCorrect := false
			for _, opt := range q.Options {
				if opt.Text == "" {
					return fmt.Errorf("option in question %d cannot be empty", i+1)
				}
				if opt.IsCorrect {
					hasCorrect = true
				}
			}
			if !hasCorrect {
				return fmt.Errorf("question %d must have at least one correct answer", i+1)
			}
		}
	}
	return nil
}

func (s *Service) CreateQuiz(ctx context.Context, quiz *models.Quiz) error {
	if err := s.validateQuiz(quiz); err != nil {
		return err
	}
	return s.repo.CreateQuiz(ctx, quiz)
}

func (s *Service) UpdateQuiz(ctx context.Context, quiz *models.Quiz) error {
	existing, err := s.repo.GetFullQuiz(ctx, quiz.ID)
	if err != nil {
		return err
	}
	if existing.TeacherID != quiz.TeacherID {
		return errors.New("forbidden")
	}
	if err := s.validateQuiz(quiz); err != nil {
		return err
	}
	return s.repo.UpdateQuiz(ctx, quiz)
}

func (s *Service) GetTeacherQuizzes(ctx context.Context, teacherID uuid.UUID, page, pageSize int) ([]models.Quiz, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize
	return s.repo.GetQuizzesByTeacher(ctx, teacherID, pageSize, offset)
}

func (s *Service) GetPublicQuizzes(ctx context.Context, page, pageSize int) ([]models.Quiz, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize
	return s.repo.GetPublicQuizzes(ctx, pageSize, offset)
}

func (s *Service) GetQuizFull(ctx context.Context, id int64) (*models.Quiz, error) {
	return s.repo.GetFullQuiz(ctx, id)
}

func (s *Service) DeleteQuiz(ctx context.Context, id int64, teacherID uuid.UUID) error {
	return s.repo.DeleteQuiz(ctx, id, teacherID)
}

func (s *Service) SaveSessionReport(ctx context.Context, session *models.GameSession) error {
	return s.repo.SaveGameSession(ctx, session)
}

func (s *Service) GetReports(ctx context.Context, teacherID uuid.UUID) ([]models.GameSession, error) {
	return s.repo.GetSessionsByTeacher(ctx, teacherID)
}

func (s *Service) GetReportExportData(ctx context.Context, id uuid.UUID) (*models.GameSession, error) {
	return s.repo.GetSessionByID(ctx, id)
}
