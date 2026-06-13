package repository

import (
	"context"
	"time"

	"phototransferlan/backend/internal/domain/entity"
)

type UserRepository interface {
	Count(ctx context.Context) (int64, error)
	Create(ctx context.Context, user *entity.User) error
	FindByUsername(ctx context.Context, username string) (*entity.User, error)
}

type UploadRepository interface {
	Create(ctx context.Context, upload *entity.Upload) error
	FindByID(ctx context.Context, id string) (*entity.Upload, error)
	FindBySHA256(ctx context.Context, sha string) (*entity.Upload, error)
	Update(ctx context.Context, upload *entity.Upload) error
	StatsForToday(ctx context.Context) (files int64, bytes int64, err error)
}

type SettingsRepository interface {
	Get(ctx context.Context) (*entity.Settings, error)
	Save(ctx context.Context, settings *entity.Settings) error
}

type SessionRepository interface {
	Create(ctx context.Context, session *entity.Session) error
	FindByID(ctx context.Context, id string) (*entity.Session, error)
	Delete(ctx context.Context, id string) error
	DeleteExpired(ctx context.Context, now time.Time) error
}

type TemporaryTokenRepository interface {
	Create(ctx context.Context, token *entity.TemporaryToken) error
	FindByHash(ctx context.Context, hash string) (*entity.TemporaryToken, error)
	Delete(ctx context.Context, id string) error
	DeleteExpired(ctx context.Context, now time.Time) error
}

type ActivityLogRepository interface {
	Create(ctx context.Context, log *entity.ActivityLog) error
	List(ctx context.Context, limit int) ([]entity.ActivityLog, error)
}
