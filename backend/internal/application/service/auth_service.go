package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"phototransferlan/backend/internal/domain/entity"
	"phototransferlan/backend/internal/domain/repository"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrSetupComplete      = errors.New("setup already completed")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrValidation         = errors.New("validation failed")
)

type AuthService struct {
	users repository.UserRepository
}

func NewAuthService(users repository.UserRepository) *AuthService {
	return &AuthService{users: users}
}

func (s *AuthService) Setup(ctx context.Context, username string, password string) error {
	count, err := s.users.Count(ctx)
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrSetupComplete
	}
	if len(strings.TrimSpace(username)) < 3 || len(password) < 8 {
		return ErrValidation
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.users.Create(ctx, &entity.User{
		Username:     strings.TrimSpace(username),
		PasswordHash: string(hash),
		CreatedAt:    time.Now(),
	})
}

func (s *AuthService) Login(ctx context.Context, username string, password string) error {
	user, err := s.users.FindByUsername(ctx, strings.TrimSpace(username))
	if err != nil {
		return ErrInvalidCredentials
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return ErrInvalidCredentials
	}
	return nil
}

func (s *AuthService) LoginUser(ctx context.Context, username string, password string) (*entity.User, error) {
	user, err := s.users.FindByUsername(ctx, strings.TrimSpace(username))
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return nil, ErrInvalidCredentials
	}
	return user, nil
}
