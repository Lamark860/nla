package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"nla/internal/model"
)

// mockUserRepo implements UserRepo for testing
type mockUserRepo struct {
	users  map[string]*model.User
	nextID int64
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		users:  make(map[string]*model.User),
		nextID: 1,
	}
}

func (m *mockUserRepo) Create(ctx context.Context, user *model.User) error {
	if _, exists := m.users[user.Email]; exists {
		return errors.New("duplicate email")
	}
	user.ID = m.nextID
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	m.nextID++
	m.users[user.Email] = user
	return nil
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	user, ok := m.users[email]
	if !ok {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (m *mockUserRepo) GetByID(ctx context.Context, id int64) (*model.User, error) {
	for _, u := range m.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, errors.New("user not found")
}

func TestRegister_Success(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewAuthService(repo, "test-secret", 72)

	resp, err := svc.Register(context.Background(), model.RegisterRequest{
		Email:    "user@example.com",
		Password: "password123",
		Name:     "Test User",
	})

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if resp.Token == "" {
		t.Error("expected token to be non-empty")
	}
	if resp.User.Email != "user@example.com" {
		t.Errorf("expected email user@example.com, got %s", resp.User.Email)
	}
	if resp.User.Name != "Test User" {
		t.Errorf("expected name Test User, got %s", resp.User.Name)
	}
	if resp.User.ID == 0 {
		t.Error("expected user ID to be set")
	}
}

func TestRegister_EmptyEmail(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewAuthService(repo, "test-secret", 72)

	_, err := svc.Register(context.Background(), model.RegisterRequest{
		Email:    "",
		Password: "password123",
	})

	if err == nil {
		t.Fatal("expected error for empty email")
	}
	if err.Error() != "email and password are required" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRegister_ShortPassword(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewAuthService(repo, "test-secret", 72)

	_, err := svc.Register(context.Background(), model.RegisterRequest{
		Email:    "user@example.com",
		Password: "short",
	})

	if err == nil {
		t.Fatal("expected error for short password")
	}
	if err.Error() != "password must be at least 8 characters" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewAuthService(repo, "test-secret", 72)

	_, err := svc.Register(context.Background(), model.RegisterRequest{
		Email:    "user@example.com",
		Password: "password123",
		Name:     "First",
	})
	if err != nil {
		t.Fatalf("first register failed: %v", err)
	}

	_, err = svc.Register(context.Background(), model.RegisterRequest{
		Email:    "user@example.com",
		Password: "password456",
		Name:     "Second",
	})
	if err == nil {
		t.Fatal("expected error for duplicate email")
	}
}

func TestLogin_Success(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewAuthService(repo, "test-secret", 72)

	// Register first
	_, err := svc.Register(context.Background(), model.RegisterRequest{
		Email:    "user@example.com",
		Password: "password123",
		Name:     "Test",
	})
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	// Login
	resp, err := svc.Login(context.Background(), model.LoginRequest{
		Email:    "user@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if resp.Token == "" {
		t.Error("expected token to be non-empty")
	}
	if resp.User.Email != "user@example.com" {
		t.Errorf("expected email user@example.com, got %s", resp.User.Email)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewAuthService(repo, "test-secret", 72)

	_, _ = svc.Register(context.Background(), model.RegisterRequest{
		Email:    "user@example.com",
		Password: "password123",
		Name:     "Test",
	})

	_, err := svc.Login(context.Background(), model.LoginRequest{
		Email:    "user@example.com",
		Password: "wrongpassword",
	})
	if err == nil {
		t.Fatal("expected error for wrong password")
	}
	if err.Error() != "invalid credentials" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestLogin_NonExistentUser(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewAuthService(repo, "test-secret", 72)

	_, err := svc.Login(context.Background(), model.LoginRequest{
		Email:    "nobody@example.com",
		Password: "password123",
	})
	if err == nil {
		t.Fatal("expected error for non-existent user")
	}
	if err.Error() != "invalid credentials" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestLogin_EmptyFields(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewAuthService(repo, "test-secret", 72)

	_, err := svc.Login(context.Background(), model.LoginRequest{})
	if err == nil {
		t.Fatal("expected error for empty fields")
	}
}

func TestGetUser_Success(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewAuthService(repo, "test-secret", 72)

	regResp, _ := svc.Register(context.Background(), model.RegisterRequest{
		Email:    "user@example.com",
		Password: "password123",
		Name:     "Test",
	})

	user, err := svc.GetUser(context.Background(), regResp.User.ID)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if user.Email != "user@example.com" {
		t.Errorf("expected email user@example.com, got %s", user.Email)
	}
}

func TestGetUser_NotFound(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewAuthService(repo, "test-secret", 72)

	_, err := svc.GetUser(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for non-existent user")
	}
}

func TestGenerateToken_ValidJWT(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewAuthService(repo, "test-secret", 72)

	resp, err := svc.Register(context.Background(), model.RegisterRequest{
		Email:    "user@example.com",
		Password: "password123",
		Name:     "Test",
	})
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	// Token should be a valid JWT with 3 segments
	parts := 0
	for _, c := range resp.Token {
		if c == '.' {
			parts++
		}
	}
	if parts != 2 {
		t.Errorf("expected JWT with 2 dots, got %d", parts)
	}
}

func TestRegister_PasswordIsHashed(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewAuthService(repo, "test-secret", 72)

	_, err := svc.Register(context.Background(), model.RegisterRequest{
		Email:    "user@example.com",
		Password: "password123",
		Name:     "Test",
	})
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	stored := repo.users["user@example.com"]
	if stored.Password == "password123" {
		t.Error("password should be hashed, not stored in plain text")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(stored.Password), []byte("password123")); err != nil {
		t.Error("stored password should be valid bcrypt hash")
	}
}
