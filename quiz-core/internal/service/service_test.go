package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"quiz-core/internal/config"
	"quiz-core/internal/models"
	"quiz-core/pkg/auth"

	"github.com/google/uuid"
)

type MockRepo struct {
	CreateQuizFunc     func(ctx context.Context, quiz *models.Quiz) error
	GetUserByIDFunc    func(ctx context.Context, id uuid.UUID) (*models.User, error)
	UpdatePasswordFunc func(ctx context.Context, userID uuid.UUID, hash string) error
}

func (m *MockRepo) CreateUser(ctx context.Context, user *models.User) error { return nil }
func (m *MockRepo) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	return nil, nil
}
func (m *MockRepo) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	if m.GetUserByIDFunc != nil {
		return m.GetUserByIDFunc(ctx, id)
	}
	return nil, nil
}
func (m *MockRepo) UpdatePassword(ctx context.Context, userID uuid.UUID, hash string) error {
	if m.UpdatePasswordFunc != nil {
		return m.UpdatePasswordFunc(ctx, userID, hash)
	}
	return nil
}
func (m *MockRepo) SetResetToken(ctx context.Context, userID uuid.UUID, token string, expiry time.Time) error {
	return nil
}
func (m *MockRepo) GetUserByResetToken(ctx context.Context, token string) (*models.User, error) {
	return nil, nil
}
func (m *MockRepo) ResetPassword(ctx context.Context, userID uuid.UUID, newHash string) error {
	return nil
}
func (m *MockRepo) CreateQuiz(ctx context.Context, quiz *models.Quiz) error {
	if m.CreateQuizFunc != nil {
		return m.CreateQuizFunc(ctx, quiz)
	}
	return nil
}
func (m *MockRepo) UpdateQuiz(ctx context.Context, quiz *models.Quiz) error { return nil }
func (m *MockRepo) GetQuizzesByTeacher(ctx context.Context, teacherID uuid.UUID, limit, offset int) ([]models.Quiz, error) {
	return nil, nil
}
func (m *MockRepo) GetPublicQuizzes(ctx context.Context, limit, offset int) ([]models.Quiz, error) {
	return nil, nil
}
func (m *MockRepo) GetFullQuiz(ctx context.Context, id int64) (*models.Quiz, error)     { return nil, nil }
func (m *MockRepo) DeleteQuiz(ctx context.Context, id int64, teacherID uuid.UUID) error { return nil }
func (m *MockRepo) SaveGameSession(ctx context.Context, session *models.GameSession) error {
	return nil
}
func (m *MockRepo) GetSessionsByTeacher(ctx context.Context, teacherID uuid.UUID) ([]models.GameSession, error) {
	return nil, nil
}
func (m *MockRepo) GetSessionByID(ctx context.Context, id uuid.UUID) (*models.GameSession, error) {
	return nil, nil
}

func TestCreateQuiz_Validation(t *testing.T) {
	mockRepo := &MockRepo{}
	svc := NewService(mockRepo, &config.Config{})

	tests := []struct {
		name    string
		quiz    *models.Quiz
		wantErr bool
	}{
		{
			name:    "Empty Title",
			quiz:    &models.Quiz{Title: "", Questions: []models.Question{{Text: "Q1"}}},
			wantErr: true,
		},
		{
			name:    "No Questions",
			quiz:    &models.Quiz{Title: "Title", Questions: []models.Question{}},
			wantErr: true,
		},
		{
			name: "Single Choice without correct answer",
			quiz: &models.Quiz{
				Title: "Title",
				Questions: []models.Question{
					{
						Type: "single_choice",
						Text: "Q1",
						Options: []models.Option{
							{Text: "A", IsCorrect: false},
							{Text: "B", IsCorrect: false},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Valid Quiz",
			quiz: &models.Quiz{
				Title: "Title",
				Questions: []models.Question{
					{
						Type: "single_choice",
						Text: "Q1",
						Options: []models.Option{
							{Text: "A", IsCorrect: true},
							{Text: "B", IsCorrect: false},
						},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.CreateQuiz(context.Background(), tt.quiz)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateQuiz() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestChangePassword(t *testing.T) {
	oldPass := "oldpass"
	oldHash, _ := auth.HashPassword(oldPass)
	userID := uuid.New()

	mockRepo := &MockRepo{
		GetUserByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.User, error) {
			return &models.User{ID: userID, PasswordHash: oldHash}, nil
		},
		UpdatePasswordFunc: func(ctx context.Context, uid uuid.UUID, hash string) error {
			if uid != userID {
				return errors.New("wrong id")
			}
			return nil
		},
	}
	svc := NewService(mockRepo, &config.Config{})

	err := svc.ChangePassword(context.Background(), userID, oldPass, "newpass")
	if err != nil {
		t.Errorf("ChangePassword failed: %v", err)
	}

	err = svc.ChangePassword(context.Background(), userID, "wrongpass", "newpass")
	if err == nil {
		t.Error("Expected error for wrong password, got nil")
	}
}
