package service

import (
	"context"
	"errors"
	"testing"

	"phototransferlan/backend/internal/domain/entity"
)

type memoryUsers struct {
	items map[string]*entity.User
}

func (m *memoryUsers) Count(context.Context) (int64, error) { return int64(len(m.items)), nil }
func (m *memoryUsers) Create(_ context.Context, user *entity.User) error {
	m.items[user.Username] = user
	return nil
}
func (m *memoryUsers) FindByUsername(_ context.Context, username string) (*entity.User, error) {
	user, ok := m.items[username]
	if !ok {
		return nil, errors.New("not found")
	}
	return user, nil
}

func TestSetupAndLogin(t *testing.T) {
	users := &memoryUsers{items: map[string]*entity.User{}}
	auth := NewAuthService(users)
	if err := auth.Setup(context.Background(), "owner", "strongpass"); err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	if err := auth.Login(context.Background(), "owner", "strongpass"); err != nil {
		t.Fatalf("login failed: %v", err)
	}
	if err := auth.Login(context.Background(), "owner", "wrongpass"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected invalid credentials, got %v", err)
	}
}
