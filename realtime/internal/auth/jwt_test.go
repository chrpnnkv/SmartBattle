package auth_test

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/chrpnnkv/SmartBattle/backend-realtime/internal/auth"
)

const testSecret = "test_jwt_secret_12345"

// makeToken создаёт тестовый JWT-токен.
func makeToken(t *testing.T, userID, role string, expiresIn time.Duration) string {
	t.Helper()
	claims := &auth.Claims{
		UserID: userID,
		Email:  "test@example.com",
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	str, err := token.SignedString([]byte(testSecret))
	if err != nil {
		t.Fatalf("не удалось создать токен: %v", err)
	}
	return str
}

// TestVerifyValidTeacherToken проверяет валидный токен преподавателя.
func TestVerifyValidTeacherToken(t *testing.T) {
	svc := auth.NewService(testSecret)
	tokenStr := makeToken(t, "user-123", "teacher", time.Hour)

	claims, err := svc.Verify(tokenStr)
	if err != nil {
		t.Fatalf("ошибка верификации токена: %v", err)
	}
	if claims.UserID != "user-123" {
		t.Errorf("неверный UserID: %s", claims.UserID)
	}
	if claims.Role != "teacher" {
		t.Errorf("неверная роль: %s", claims.Role)
	}
	if !claims.IsTeacher() {
		t.Error("IsTeacher() должен возвращать true")
	}
}

// TestVerifyValidStudentToken проверяет валидный токен студента.
func TestVerifyValidStudentToken(t *testing.T) {
	svc := auth.NewService(testSecret)
	tokenStr := makeToken(t, "student-456", "student", time.Hour)

	claims, err := svc.Verify(tokenStr)
	if err != nil {
		t.Fatalf("ошибка верификации токена: %v", err)
	}
	if claims.IsTeacher() {
		t.Error("IsTeacher() должен возвращать false для студента")
	}
}

// TestVerifyExpiredToken проверяет отклонение истёкшего токена.
func TestVerifyExpiredToken(t *testing.T) {
	svc := auth.NewService(testSecret)
	tokenStr := makeToken(t, "user-1", "teacher", -time.Hour)

	_, err := svc.Verify(tokenStr)
	if err == nil {
		t.Error("истёкший токен должен быть отклонён")
	}
}

// TestVerifyWrongSecret проверяет отклонение токена с неверным секретом.
func TestVerifyWrongSecret(t *testing.T) {
	svc := auth.NewService("wrong_secret")
	tokenStr := makeToken(t, "user-1", "teacher", time.Hour)

	_, err := svc.Verify(tokenStr)
	if err == nil {
		t.Error("токен с неверным секретом должен быть отклонён")
	}
}

// TestVerifyEmptyToken проверяет поведение с пустым токеном.
func TestVerifyEmptyToken(t *testing.T) {
	svc := auth.NewService(testSecret)
	_, err := svc.Verify("")
	if err == nil {
		t.Error("пустой токен должен быть отклонён")
	}
}

// TestVerifyMalformedToken проверяет поведение с некорректным токеном.
func TestVerifyMalformedToken(t *testing.T) {
	svc := auth.NewService(testSecret)
	_, err := svc.Verify("not.a.valid.jwt.token")
	if err == nil {
		t.Error("некорректный токен должен быть отклонён")
	}
}
