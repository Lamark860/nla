package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "test-jwt-secret"

func generateTestToken(userID int64, secret string, expDelta time.Duration) string {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(expDelta).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, _ := token.SignedString([]byte(secret))
	return s
}

func TestJWT_ValidToken(t *testing.T) {
	token := generateTestToken(42, testSecret, time.Hour)

	handler := JWT(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := GetUserID(r.Context())
		if !ok {
			t.Error("expected user_id in context")
		}
		if userID != 42 {
			t.Errorf("expected user_id 42, got %d", userID)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestJWT_MissingHeader(t *testing.T) {
	handler := JWT(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestJWT_InvalidFormat(t *testing.T) {
	tests := []struct {
		name   string
		header string
	}{
		{"no bearer prefix", "Token abc123"},
		{"only bearer", "Bearer"},
		{"empty value", "Bearer "},
		{"basic auth", "Basic dXNlcjpwYXNz"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := JWT(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				t.Error("handler should not be called")
			}))

			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Authorization", tt.header)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Errorf("expected 401, got %d", rec.Code)
			}
		})
	}
}

func TestJWT_ExpiredToken(t *testing.T) {
	token := generateTestToken(1, testSecret, -time.Hour)

	handler := JWT(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestJWT_WrongSecret(t *testing.T) {
	token := generateTestToken(1, "other-secret", time.Hour)

	handler := JWT(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestJWT_AlgorithmConfusion(t *testing.T) {
	// Try signing with 'none' algorithm — should be rejected
	token := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{
		"user_id": float64(1),
		"exp":     time.Now().Add(time.Hour).Unix(),
	})
	s, _ := token.SignedString(jwt.UnsafeAllowNoneSignatureType)

	handler := JWT(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called with none algorithm")
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+s)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestGetUserID_NoContext(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	id, ok := GetUserID(req.Context())
	if ok {
		t.Error("expected ok=false when no user in context")
	}
	if id != 0 {
		t.Errorf("expected id=0, got %d", id)
	}
}

func TestJWT_MalformedToken(t *testing.T) {
	handler := JWT(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer not.a.valid.jwt.token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestJWT_MissingUserIDClaim(t *testing.T) {
	claims := jwt.MapClaims{
		"exp": time.Now().Add(time.Hour).Unix(),
		// no user_id
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, _ := token.SignedString([]byte(testSecret))

	handler := JWT(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called without user_id claim")
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}
