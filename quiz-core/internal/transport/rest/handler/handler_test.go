package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"quiz-core/internal/config"
	"quiz-core/internal/models"
	"quiz-core/internal/service"
	"quiz-core/pkg/auth"

	"github.com/google/uuid"
)

type MockRepo struct {
	CreateUserFunc     func(ctx context.Context, user *models.User) error
	GetUserByEmailFunc func(ctx context.Context, email string) (*models.User, error)
}

func (m *MockRepo) CreateUser(ctx context.Context, user *models.User) error {
	if m.CreateUserFunc != nil {
		return m.CreateUserFunc(ctx, user)
	}
	return nil
}
func (m *MockRepo) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	if m.GetUserByEmailFunc != nil {
		return m.GetUserByEmailFunc(ctx, email)
	}
	return nil, nil
}

func (m *MockRepo) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	return nil, nil
}
func (m *MockRepo) UpdatePassword(ctx context.Context, userID uuid.UUID, hash string) error {
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
func (m *MockRepo) CreateQuiz(ctx context.Context, quiz *models.Quiz) error { return nil }
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

func TestRegisterHandler(t *testing.T) {
	mockRepo := &MockRepo{
		CreateUserFunc: func(ctx context.Context, user *models.User) error {
			return nil // Успешное создание
		},
	}
	svc := service.NewService(mockRepo, &config.Config{JWTSecret: "secret"})
	h := NewHandler(svc, &config.Config{})

	body := []byte(`{"email":"test@test.com", "password":"pass"}`)
	req := httptest.NewRequest("POST", "/auth/register", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	h.Register(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, http.StatusCreated)
	}
}

func TestLoginHandler(t *testing.T) {
	hashedPass, _ := auth.HashPassword("pass")
	mockRepo := &MockRepo{
		GetUserByEmailFunc: func(ctx context.Context, email string) (*models.User, error) {
			return &models.User{
				ID:           uuid.New(),
				Email:        email,
				PasswordHash: hashedPass,
				Role:         "teacher",
			}, nil
		},
	}
	svc := service.NewService(mockRepo, &config.Config{JWTSecret: "secret"})
	h := NewHandler(svc, &config.Config{})

	body := []byte(`{"email":"test@test.com", "password":"pass"}`)
	req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	h.Login(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}

	var resp map[string]string
	json.NewDecoder(rr.Body).Decode(&resp)
	if _, ok := resp["token"]; !ok {
		t.Error("response does not contain token")
	}
}
