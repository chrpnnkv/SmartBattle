package service

import (
	"errors"
	"log"
	"time"

	"github.com/chrpnnkv/SmartBattle/internal/config"
	"github.com/chrpnnkv/SmartBattle/internal/models"
	"github.com/chrpnnkv/SmartBattle/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo *repository.UserRepository
	cfg  *config.Config
}

func NewAuthService(repo *repository.UserRepository, cfg *config.Config) *AuthService {
	return &AuthService{repo: repo, cfg: cfg}
}

func (s *AuthService) Register(email, password string) (*models.User, error) {
	existing, err := s.repo.GetByEmail(email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("user already exists")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Email:        email,
		PasswordHash: string(hash),
		Role:         "teacher",
	}

	if err := s.repo.Create(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *AuthService) Login(email, password string) (string, error) {
	user, err := s.repo.GetByEmail(email)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID.String(),
		"role":    user.Role,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *AuthService) ChangePassword(userID uuid.UUID, oldPassword, newPassword string) error {
	user, err := s.repo.GetByID(userID)
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return errors.New("invalid old password")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.PasswordHash = string(hash)
	return s.repo.Update(user)
}

func (s *AuthService) ForgotPassword(email string) error {
	user, err := s.repo.GetByEmail(email)
	if err != nil || user == nil {
		return errors.New("user not found")
	}

	token := uuid.New().String()
	expiry := time.Now().Add(time.Hour * 1)

	user.ResetToken = &token
	user.ResetTokenExpiry = &expiry

	if err := s.repo.Update(user); err != nil {
		return err
	}

	log.Printf("RESET PASSWORD TOKEN FOR %s: %s\n", email, token)
	return nil
}

func (s *AuthService) ResetPassword(token, newPassword string) error {
	user, err := s.repo.GetByResetToken(token)
	if err != nil {
		return errors.New("invalid or expired token")
	}

	if user.ResetTokenExpiry == nil || user.ResetTokenExpiry.Before(time.Now()) {
		return errors.New("token expired")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.PasswordHash = string(hash)
	user.ResetToken = nil
	user.ResetTokenExpiry = nil

	return s.repo.Update(user)
}
