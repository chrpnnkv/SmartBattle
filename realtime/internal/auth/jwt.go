package auth

import (
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// Service — сервис работы с JWT
type Service struct {
	secret []byte
}

// NewService создаёт новый JWT-сервис.
func NewService(secret string) *Service {
	return &Service{secret: []byte(secret)}
}

// Verify разбирает и проверяет JWT-токен.
// Возвращает Claims при успехе или ошибку при невалидном токене.
func (s *Service) Verify(tokenStr string) (*Claims, error) {
	if tokenStr == "" {
		return nil, errors.New("токен не передан")
	}

	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("неожиданный алгоритм подписи: %v", t.Header["alg"])
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("невалидный токен: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("невалидные claims в токене")
	}

	return claims, nil
}

// IsTeacher проверяет, является ли пользователь преподавателем.
func (c *Claims) IsTeacher() bool {
	return c.Role == "teacher"
}
