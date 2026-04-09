package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"nla/internal/middleware"
	"nla/internal/model"
)

// setupTestRouter creates a minimal chi router for handler tests (no circular import)
func setupTestRouter(t *testing.T) http.Handler {
	t.Helper()
	repo := newMockUserRepo()
	svc := newTestAuthService(repo)
	h := New(svc, "test-secret")

	r := chi.NewRouter()
	r.Get("/health", h.Health)
	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/auth/register", h.Register)
		r.Post("/auth/login", h.Login)
		r.Group(func(r chi.Router) {
			r.Use(middleware.JWT(h.JWTSecret))
			r.Get("/auth/me", h.Me)
		})
	})
	return r
}

func TestHealth_Endpoint(t *testing.T) {
	r := setupTestRouter(t)

	req := httptest.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var body map[string]string
	json.NewDecoder(rec.Body).Decode(&body)
	if body["status"] != "ok" {
		t.Errorf("expected status=ok, got %s", body["status"])
	}
}

func TestRegister_Handler_Success(t *testing.T) {
	r := setupTestRouter(t)

	payload, _ := json.Marshal(model.RegisterRequest{
		Email:    "new@example.com",
		Password: "password123",
		Name:     "New User",
	})

	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp model.AuthResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Token == "" {
		t.Error("expected token in response")
	}
	if resp.User.Email != "new@example.com" {
		t.Errorf("expected email new@example.com, got %s", resp.User.Email)
	}
}

func TestRegister_Handler_InvalidJSON(t *testing.T) {
	r := setupTestRouter(t)

	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestRegister_Handler_ValidationError(t *testing.T) {
	r := setupTestRouter(t)

	payload, _ := json.Marshal(model.RegisterRequest{
		Email:    "user@example.com",
		Password: "short",
	})

	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestLogin_Handler_Success(t *testing.T) {
	r := setupTestRouter(t)

	// Register
	regPayload, _ := json.Marshal(model.RegisterRequest{
		Email:    "user@example.com",
		Password: "password123",
		Name:     "Test",
	})
	regReq := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewReader(regPayload))
	regReq.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(httptest.NewRecorder(), regReq)

	// Login
	loginPayload, _ := json.Marshal(model.LoginRequest{
		Email:    "user@example.com",
		Password: "password123",
	})
	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(loginPayload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp model.AuthResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Token == "" {
		t.Error("expected token in response")
	}
}

func TestLogin_Handler_WrongPassword(t *testing.T) {
	r := setupTestRouter(t)

	// Register
	regPayload, _ := json.Marshal(model.RegisterRequest{
		Email:    "user@example.com",
		Password: "password123",
		Name:     "Test",
	})
	regReq := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewReader(regPayload))
	regReq.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(httptest.NewRecorder(), regReq)

	// Login with wrong password
	loginPayload, _ := json.Marshal(model.LoginRequest{
		Email:    "user@example.com",
		Password: "wrongpassword",
	})
	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(loginPayload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestMe_Handler_WithToken(t *testing.T) {
	r := setupTestRouter(t)

	// Register and get token
	regPayload, _ := json.Marshal(model.RegisterRequest{
		Email:    "user@example.com",
		Password: "password123",
		Name:     "Test User",
	})
	regReq := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewReader(regPayload))
	regReq.Header.Set("Content-Type", "application/json")
	regRec := httptest.NewRecorder()
	r.ServeHTTP(regRec, regReq)

	var regResp model.AuthResponse
	json.NewDecoder(regRec.Body).Decode(&regResp)

	// Get /me
	req := httptest.NewRequest("GET", "/api/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+regResp.Token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var user model.User
	json.NewDecoder(rec.Body).Decode(&user)
	if user.Email != "user@example.com" {
		t.Errorf("expected email user@example.com, got %s", user.Email)
	}
}

func TestMe_Handler_NoToken(t *testing.T) {
	r := setupTestRouter(t)

	req := httptest.NewRequest("GET", "/api/v1/auth/me", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestNotFound(t *testing.T) {
	r := setupTestRouter(t)

	req := httptest.NewRequest("GET", "/api/v1/nonexistent", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}
