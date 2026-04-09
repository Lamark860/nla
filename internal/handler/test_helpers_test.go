package handler

import (
	"context"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"

	"nla/internal/model"
	"nla/internal/service"
)

// mockUserRepo implements service.UserRepo for handler integration tests
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

func (m *mockUserRepo) Create(_ context.Context, user *model.User) error {
	if _, exists := m.users[user.Email]; exists {
		return errors.New("duplicate email")
	}
	user.ID = m.nextID
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	m.nextID++
	// Store a copy to avoid mutation
	stored := *user
	m.users[user.Email] = &stored
	return nil
}

func (m *mockUserRepo) GetByEmail(_ context.Context, email string) (*model.User, error) {
	user, ok := m.users[email]
	if !ok {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (m *mockUserRepo) GetByID(_ context.Context, id int64) (*model.User, error) {
	for _, u := range m.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, errors.New("user not found")
}

func newTestAuthService(repo *mockUserRepo) *service.AuthService {
	return service.NewAuthService(repo, "test-secret", 72)
}

// helper to pre-create a user in the mock repo
func seedUser(repo *mockUserRepo, email, password, name string) *model.User {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	user := &model.User{
		ID:        repo.nextID,
		Email:     email,
		Password:  string(hash),
		Name:      name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	repo.nextID++
	repo.users[email] = user
	return user
}
