package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"quiz-core/pkg/auth"

	"github.com/google/uuid"
)

func TestAuthMiddleware(t *testing.T) {
	secret := "testsecret"
	validToken, _ := auth.GenerateToken(uuid.New(), "teacher", secret)

	mw := AuthMiddleware(secret)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
	}{
		{
			name:           "No Header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Invalid Format",
			authHeader:     "Basic " + validToken,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Invalid Token",
			authHeader:     "Bearer invalid.token.string",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Valid Token",
			authHeader:     "Bearer " + validToken,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			rr := httptest.NewRecorder()

			mw(nextHandler).ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, tt.expectedStatus)
			}
		})
	}
}
