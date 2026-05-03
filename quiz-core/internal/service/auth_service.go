package service

import (
	"errors"
	"log"
	"os"
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
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID.String(),
		"role":    user.Role,
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

// ChangePassword меняет пароль и возвращает свежий JWT.
// Ротация токена на смене пароля — это (а) безопасность (старый токен формально остаётся
// валидным, новый сразу заменит его на клиенте) и (б) UX: после изменения пароля у юзера
// гарантированно полные 7 дней до следующего истечения, а не остаток от старого токена.
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

	// Логирование сырого reset-токена допустимо только в dev-окружении.
	// В production эта строка превращает доступ к stdout/логам в захват аккаунта.
	// Включается переменной окружения LOG_RESET_TOKENS=true.
	if os.Getenv("LOG_RESET_TOKENS") == "true" {
		log.Printf("RESET PASSWORD TOKEN FOR %s: %s\n", email, token)
	} else {
		log.Printf("Password reset requested for %s (token logging disabled; set LOG_RESET_TOKENS=true to see it in logs)", email)
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
