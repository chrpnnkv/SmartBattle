package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestHashPassword(t *testing.T) {
	password := "secret123"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if hash == password {
		t.Error("Hash should not match plain password")
	}

	if !CheckPassword(password, hash) {
		t.Error("CheckPassword failed for correct password")
	}

	if CheckPassword("wrong", hash) {
		t.Error("CheckPassword passed for wrong password")
	}
}

func TestJWT(t *testing.T) {
	secret := "mysecret"
	userID := uuid.New()
	role := "teacher"

	tokenString, err := GenerateToken(userID, role, secret)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	claims, err := ValidateToken(tokenString, secret)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("UserID mismatch: got %v, want %v", claims.UserID, userID)
	}
	if claims.Role != role {
		t.Errorf("Role mismatch: got %v, want %v", claims.Role, role)
	}

	_, err = ValidateToken(tokenString, "wrongsecret")
	if err == nil {
		t.Error("Expected error for wrong secret, got nil")
	}

	expiredClaims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
		},
	}
	expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims)
	expiredString, _ := expiredToken.SignedString([]byte(secret))

	_, err = ValidateToken(expiredString, secret)
	if err == nil {
		t.Error("Expected error for expired token, got nil")
	}
}
