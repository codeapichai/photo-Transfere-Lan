package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"phototransferlan/backend/internal/domain/entity"
	"phototransferlan/backend/internal/domain/repository"
)

var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
)

type SecurityService struct {
	sessions repository.SessionRepository
	tokens   repository.TemporaryTokenRepository
	settings repository.SettingsRepository
	logs     repository.ActivityLogRepository
}

type LoginSession struct {
	SessionID string    `json:"session_id"`
	CSRFToken string    `json:"csrf_token"`
	ExpiresAt time.Time `json:"expires_at"`
}

type TemporaryTokenResult struct {
	Token     string    `json:"token"`
	UploadURL string    `json:"upload_url"`
	ExpiresAt time.Time `json:"expires_at"`
}

func NewSecurityService(sessions repository.SessionRepository, tokens repository.TemporaryTokenRepository, settings repository.SettingsRepository, logs repository.ActivityLogRepository) *SecurityService {
	return &SecurityService{sessions: sessions, tokens: tokens, settings: settings, logs: logs}
}

func (s *SecurityService) CreateSession(ctx context.Context, userID uint) (*LoginSession, error) {
	settings, err := s.settings.Get(ctx)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	sessionID, err := randomToken(32)
	if err != nil {
		return nil, err
	}
	csrf, err := randomToken(32)
	if err != nil {
		return nil, err
	}
	expiresAt := now.Add(time.Duration(settings.SessionTimeoutMinutes) * time.Minute)
	if err := s.sessions.Create(ctx, &entity.Session{ID: sessionID, UserID: userID, CSRFToken: csrf, ExpiresAt: expiresAt, CreatedAt: now}); err != nil {
		return nil, err
	}
	_ = s.logs.Create(ctx, &entity.ActivityLog{EventType: "login", Actor: "user", Message: "User logged in", CreatedAt: now})
	return &LoginSession{SessionID: sessionID, CSRFToken: csrf, ExpiresAt: expiresAt}, nil
}

func (s *SecurityService) ValidateSession(ctx context.Context, sessionID string) (*entity.Session, error) {
	if sessionID == "" {
		return nil, ErrUnauthorized
	}
	_ = s.sessions.DeleteExpired(ctx, time.Now())
	session, err := s.sessions.FindByID(ctx, sessionID)
	if err != nil || session.ExpiresAt.Before(time.Now()) {
		return nil, ErrUnauthorized
	}
	return session, nil
}

func (s *SecurityService) ValidateCSRF(session *entity.Session, token string) error {
	if session == nil || token == "" || token != session.CSRFToken {
		return ErrForbidden
	}
	return nil
}

func (s *SecurityService) Logout(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return nil
	}
	_ = s.logs.Create(ctx, &entity.ActivityLog{EventType: "logout", Actor: "user", Message: "User logged out", CreatedAt: time.Now()})
	return s.sessions.Delete(ctx, sessionID)
}

func (s *SecurityService) CreateTemporaryToken(ctx context.Context, uploadBaseURL string) (*TemporaryTokenResult, error) {
	raw, err := randomToken(32)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	expiresAt := now.Add(15 * time.Minute)
	token := &entity.TemporaryToken{ID: hashToken(raw), TokenHash: hashToken(raw), ExpiresAt: expiresAt, CreatedAt: now}
	if err := s.tokens.Create(ctx, token); err != nil {
		return nil, err
	}
	_ = s.logs.Create(ctx, &entity.ActivityLog{EventType: "token", Actor: "user", Message: "Temporary upload token created", CreatedAt: now})
	return &TemporaryTokenResult{Token: raw, UploadURL: uploadBaseURL + "?token=" + raw, ExpiresAt: expiresAt}, nil
}

func (s *SecurityService) ValidateTemporaryToken(ctx context.Context, raw string) error {
	if raw == "" {
		return ErrUnauthorized
	}
	_ = s.tokens.DeleteExpired(ctx, time.Now())
	token, err := s.tokens.FindByHash(ctx, hashToken(raw))
	if err != nil || token.ExpiresAt.Before(time.Now()) {
		return ErrUnauthorized
	}
	return nil
}

func randomToken(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
