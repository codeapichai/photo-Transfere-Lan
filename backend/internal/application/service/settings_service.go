package service

import (
	"context"
	"time"

	"phototransferlan/backend/internal/domain/entity"
	"phototransferlan/backend/internal/domain/repository"
)

type SettingsService struct {
	settings repository.SettingsRepository
	logs     repository.ActivityLogRepository
}

func NewSettingsService(settings repository.SettingsRepository, logs repository.ActivityLogRepository) *SettingsService {
	return &SettingsService{settings: settings, logs: logs}
}

func (s *SettingsService) Get(ctx context.Context) (*entity.Settings, error) {
	return s.settings.Get(ctx)
}

func (s *SettingsService) Save(ctx context.Context, settings *entity.Settings) (*entity.Settings, error) {
	if settings.UploadDirectory == "" || settings.MaxConcurrentUploads < 1 || settings.SessionTimeoutMinutes < 5 {
		return nil, ErrValidation
	}
	settings.ID = 1
	if err := s.settings.Save(ctx, settings); err != nil {
		return nil, err
	}
	_ = s.logs.Create(ctx, &entity.ActivityLog{EventType: "settings", Actor: "user", Message: "Settings updated", CreatedAt: time.Now()})
	return settings, nil
}

type LogService struct {
	logs repository.ActivityLogRepository
}

func NewLogService(logs repository.ActivityLogRepository) *LogService {
	return &LogService{logs: logs}
}

func (s *LogService) List(ctx context.Context, limit int) ([]entity.ActivityLog, error) {
	return s.logs.List(ctx, limit)
}
