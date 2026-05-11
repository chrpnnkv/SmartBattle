package service

import (
	"errors"
	"log"
	"os"
	"time"

	"github.com/chrpnnkv/SmartBattle/internal/admins"
	"github.com/chrpnnkv/SmartBattle/internal/config"
	"github.com/chrpnnkv/SmartBattle/internal/models"
	"github.com/chrpnnkv/SmartBattle/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo         *repository.UserRepository
	cfg          *config.Config
	admins       *admins.List
	emailService EmailService
}

func NewAuthService(repo *repository.UserRepository, cfg *config.Config, adminList *admins.List, emailService EmailService) *AuthService {
	return &AuthService{
		repo:         repo,
		cfg:          cfg,
		admins:       adminList,
		emailService: emailService,
	}
}

func (s *AuthService) GetUserByID(id uuid.UUID) (*models.User, error) {
	return s.repo.GetByID(id)
}

func (s *AuthService) Register(name, email, password string) (*models.User, string, error) {
	existing, err := s.repo.GetByEmail(email)
	if err != nil {
		return nil, "", err
	}
	if existing != nil {
		return nil, "", errors.New("user already exists")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}

	user := &models.User{
		Name:         name,
		Email:        email,
		PasswordHash: string(hash),
		Role:         "teacher",
	}

	if err := s.repo.Create(user); err != nil {
		return nil, "", err
	}

	token, err := s.issueToken(user)
	if err != nil {
		return nil, "", err
	}
	return user, token, nil
}

func (s *AuthService) issueToken(user *models.User) (string, error) {
	role := user.Role
	if s.admins != nil && s.admins.IsAdmin(user.Email) {
		role = admins.RoleAdmin
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID.String(),
		"role":    role,
		"email":   user.Email,
		// 7 дней — компромисс между удобством преподавателя и безопасностью.
		"exp": time.Now().Add(time.Hour * 24 * 7).Unix(),
	})
	return token.SignedString([]byte(s.cfg.JWTSecret))
}

func (s *AuthService) Login(email, password string) (*models.User, string, error) {
	user, err := s.repo.GetByEmail(email)
	if err != nil {
		return nil, "", err
	}
	if user == nil {
		return nil, "", errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, "", errors.New("invalid credentials")
	}

	tokenString, err := s.issueToken(user)
	if err != nil {
		return nil, "", err
	}

	return user, tokenString, nil
}

func (s *AuthService) ChangePassword(userID uuid.UUID, oldPassword, newPassword string) (*models.User, string, error) {
	user, err := s.repo.GetByID(userID)
	if err != nil {
		return nil, "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return nil, "", errors.New("invalid old password")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}

	user.PasswordHash = string(hash)
	if err := s.repo.Update(user); err != nil {
		return nil, "", err
	}

	token, err := s.issueToken(user)
	if err != nil {
		return nil, "", err
	}
	return user, token, nil
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

	if s.emailService != nil {
		go func() {
			err := s.emailService.SendPasswordResetEmail(user.Email, token)
			if err != nil {
				log.Printf("[ERROR] Failed to send reset email to %s: %v", user.Email, err)
			}
		}()
	}

	if os.Getenv("LOG_RESET_TOKENS") == "true" {
		log.Printf("RESET PASSWORD TOKEN FOR %s: %s\n", email, token)
	} else {
		log.Printf("Password reset requested for %s", email)
	}
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
