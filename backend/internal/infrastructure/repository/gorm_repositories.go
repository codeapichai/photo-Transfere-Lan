package repository

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"time"

	"phototransferlan/backend/internal/domain/entity"

	"gorm.io/gorm"
)

type GormUserRepository struct{ db *gorm.DB }

func NewGormUserRepository(db *gorm.DB) *GormUserRepository { return &GormUserRepository{db: db} }

func (r *GormUserRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	return count, r.db.WithContext(ctx).Model(&entity.User{}).Count(&count).Error
}

func (r *GormUserRepository) Create(ctx context.Context, user *entity.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *GormUserRepository) FindByUsername(ctx context.Context, username string) (*entity.User, error) {
	var user entity.User
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

type GormUploadRepository struct{ db *gorm.DB }

func NewGormUploadRepository(db *gorm.DB) *GormUploadRepository { return &GormUploadRepository{db: db} }

func (r *GormUploadRepository) Create(ctx context.Context, upload *entity.Upload) error {
	return r.db.WithContext(ctx).Create(upload).Error
}

func (r *GormUploadRepository) FindByID(ctx context.Context, id string) (*entity.Upload, error) {
	var upload entity.Upload
	if err := r.db.WithContext(ctx).First(&upload, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &upload, nil
}

func (r *GormUploadRepository) FindBySHA256(ctx context.Context, sha string) (*entity.Upload, error) {
	var upload entity.Upload
	err := r.db.WithContext(ctx).Where("sha256 = ? AND status = ?", sha, entity.UploadSuccess).First(&upload).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &upload, err
}

func (r *GormUploadRepository) Update(ctx context.Context, upload *entity.Upload) error {
	return r.db.WithContext(ctx).Save(upload).Error
}

func (r *GormUploadRepository) StatsForToday(ctx context.Context) (int64, int64, error) {
	start := time.Now().Truncate(24 * time.Hour)
	var files int64
	var bytes int64
	query := r.db.WithContext(ctx).Model(&entity.Upload{}).
		Where("upload_time >= ? AND status = ?", start, entity.UploadSuccess).
		Count(&files)
	if query.Error != nil {
		return 0, 0, query.Error
	}
	err := r.db.WithContext(ctx).Model(&entity.Upload{}).
		Where("upload_time >= ? AND status = ?", start, entity.UploadSuccess).
		Select("COALESCE(SUM(filesize), 0)").
		Scan(&bytes).Error
	return files, bytes, err
}

type GormSettingsRepository struct{ db *gorm.DB }

func NewGormSettingsRepository(db *gorm.DB) *GormSettingsRepository {
	return &GormSettingsRepository{db: db}
}

func (r *GormSettingsRepository) Get(ctx context.Context) (*entity.Settings, error) {
	var settings entity.Settings
	err := r.db.WithContext(ctx).First(&settings, "id = 1").Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		settings = entity.Settings{ID: 1, UploadDirectory: defaultUploadDirectory(), AutoOrganize: true, MaxConcurrentUploads: 3, SessionTimeoutMinutes: 30}
		err = r.Save(ctx, &settings)
	}
	return &settings, err
}

func (r *GormSettingsRepository) Save(ctx context.Context, settings *entity.Settings) error {
	settings.ID = 1
	return r.db.WithContext(ctx).Save(settings).Error
}

func defaultUploadDirectory() string {
	if home := os.Getenv("USERPROFILE"); home != "" {
		return filepath.Join(home, "Pictures", "PhotoTransferLAN")
	}
	return "Uploads"
}

type GormSessionRepository struct{ db *gorm.DB }

func NewGormSessionRepository(db *gorm.DB) *GormSessionRepository {
	return &GormSessionRepository{db: db}
}

func (r *GormSessionRepository) Create(ctx context.Context, session *entity.Session) error {
	return r.db.WithContext(ctx).Create(session).Error
}

func (r *GormSessionRepository) FindByID(ctx context.Context, id string) (*entity.Session, error) {
	var session entity.Session
	if err := r.db.WithContext(ctx).First(&session, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *GormSessionRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.Session{}, "id = ?", id).Error
}

func (r *GormSessionRepository) DeleteExpired(ctx context.Context, now time.Time) error {
	return r.db.WithContext(ctx).Delete(&entity.Session{}, "expires_at < ?", now).Error
}

type GormTemporaryTokenRepository struct{ db *gorm.DB }

func NewGormTemporaryTokenRepository(db *gorm.DB) *GormTemporaryTokenRepository {
	return &GormTemporaryTokenRepository{db: db}
}

func (r *GormTemporaryTokenRepository) Create(ctx context.Context, token *entity.TemporaryToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

func (r *GormTemporaryTokenRepository) FindByHash(ctx context.Context, hash string) (*entity.TemporaryToken, error) {
	var token entity.TemporaryToken
	if err := r.db.WithContext(ctx).First(&token, "token_hash = ?", hash).Error; err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *GormTemporaryTokenRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.TemporaryToken{}, "id = ?", id).Error
}

func (r *GormTemporaryTokenRepository) DeleteExpired(ctx context.Context, now time.Time) error {
	return r.db.WithContext(ctx).Delete(&entity.TemporaryToken{}, "expires_at < ?", now).Error
}

type GormActivityLogRepository struct{ db *gorm.DB }

func NewGormActivityLogRepository(db *gorm.DB) *GormActivityLogRepository {
	return &GormActivityLogRepository{db: db}
}

func (r *GormActivityLogRepository) Create(ctx context.Context, log *entity.ActivityLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *GormActivityLogRepository) List(ctx context.Context, limit int) ([]entity.ActivityLog, error) {
	if limit <= 0 || limit > 1000 {
		limit = 500
	}
	var logs []entity.ActivityLog
	err := r.db.WithContext(ctx).Order("created_at DESC").Limit(limit).Find(&logs).Error
	return logs, err
}
