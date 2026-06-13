package service

import (
	"context"
	"errors"
	"io"
	"path/filepath"
	"time"

	"phototransferlan/backend/internal/domain/entity"
	"phototransferlan/backend/internal/domain/repository"
	"phototransferlan/backend/internal/infrastructure/storage"
	wsapi "phototransferlan/backend/internal/presentation/websocket"

	"github.com/google/uuid"
)

const DefaultChunkSize = int64(5 * 1024 * 1024)

var (
	ErrUploadNotFound = errors.New("upload not found")
	ErrEmptyChunk     = errors.New("empty upload chunk")
	ErrIncompleteFile = errors.New("upload incomplete")
)

type UploadService struct {
	uploads  repository.UploadRepository
	settings repository.SettingsRepository
	logs     repository.ActivityLogRepository
	store    *storage.LocalStore
	hub      *wsapi.Hub
}

type CreateUploadSessionInput struct {
	Filename        string
	Filesize        int64
	DeviceName      string
	DuplicatePolicy entity.DuplicatePolicy
}

type UploadSession struct {
	ID        string `json:"id"`
	ChunkSize int64  `json:"chunk_size"`
}

func NewUploadService(uploads repository.UploadRepository, settings repository.SettingsRepository, logs repository.ActivityLogRepository, store *storage.LocalStore, hub *wsapi.Hub) *UploadService {
	return &UploadService{uploads: uploads, settings: settings, logs: logs, store: store, hub: hub}
}

func (s *UploadService) CreateSession(ctx context.Context, input CreateUploadSessionInput) (*UploadSession, error) {
	if input.Filename == "" || input.Filesize < 0 {
		return nil, ErrValidation
	}
	settings, err := s.settings.Get(ctx)
	if err != nil {
		return nil, err
	}
	if settings.MaxUploadSize != nil && input.Filesize > *settings.MaxUploadSize {
		return nil, ErrValidation
	}
	now := time.Now()
	path, err := s.store.PathFor(input.Filename, settings.AutoOrganize, now)
	if settings.UploadDirectory != "" {
		path, err = storage.NewLocalStore(settings.UploadDirectory).PathFor(input.Filename, settings.AutoOrganize, now)
	}
	if err != nil {
		return nil, err
	}
	id := uuid.NewString()
	upload := &entity.Upload{
		ID:               id,
		Filename:         filepath.Base(path),
		OriginalFilename: input.Filename,
		Filesize:         input.Filesize,
		UploadTime:       now,
		DeviceName:       input.DeviceName,
		Status:           entity.UploadPending,
		StoragePath:      path,
	}
	if err := s.uploads.Create(ctx, upload); err != nil {
		return nil, err
	}
	_ = s.logs.Create(ctx, &entity.ActivityLog{EventType: "upload", Actor: input.DeviceName, Message: "Upload session created: " + input.Filename, CreatedAt: now})
	return &UploadSession{ID: id, ChunkSize: DefaultChunkSize}, nil
}

func (s *UploadService) AppendChunk(ctx context.Context, id string, r io.Reader) error {
	upload, err := s.uploads.FindByID(ctx, id)
	if err != nil {
		return ErrUploadNotFound
	}
	written, err := s.store.Append(upload.StoragePath, r)
	if err != nil {
		upload.Status = entity.UploadFailed
		_ = s.uploads.Update(ctx, upload)
		return err
	}
	if written == 0 && upload.Filesize > upload.ReceivedBytes {
		upload.Status = entity.UploadFailed
		_ = s.uploads.Update(ctx, upload)
		return ErrEmptyChunk
	}
	upload.ReceivedBytes += written
	upload.Status = entity.UploadUploading
	if err := s.uploads.Update(ctx, upload); err != nil {
		return err
	}
	s.hub.Publish(wsapi.Event{Type: "upload_progress", Data: upload})
	return nil
}

func (s *UploadService) Complete(ctx context.Context, id string, expectedSHA string, policy entity.DuplicatePolicy) (*entity.Upload, error) {
	upload, err := s.uploads.FindByID(ctx, id)
	if err != nil {
		return nil, ErrUploadNotFound
	}
	if upload.ReceivedBytes < upload.Filesize {
		upload.Status = entity.UploadFailed
		_ = s.uploads.Update(ctx, upload)
		return upload, ErrIncompleteFile
	}
	sha, err := s.store.SHA256(upload.StoragePath)
	if err != nil {
		upload.Status = entity.UploadFailed
		_ = s.uploads.Update(ctx, upload)
		return upload, err
	}
	upload.SHA256 = sha
	if expectedSHA != "" && expectedSHA != sha {
		upload.Status = entity.UploadCorrupted
		_ = s.uploads.Update(ctx, upload)
		return upload, nil
	}
	duplicate, err := s.uploads.FindBySHA256(ctx, sha)
	if err != nil {
		return nil, err
	}
	if duplicate != nil && duplicate.ID != upload.ID && policy == entity.DuplicateSkip {
		upload.Status = entity.UploadDuplicate
	} else {
		upload.Status = entity.UploadSuccess
	}
	if err := s.uploads.Update(ctx, upload); err != nil {
		return nil, err
	}
	_ = s.logs.Create(ctx, &entity.ActivityLog{EventType: "upload", Actor: upload.DeviceName, Message: "Upload completed: " + upload.OriginalFilename + " (" + string(upload.Status) + ")", CreatedAt: time.Now()})
	s.hub.Publish(wsapi.Event{Type: "upload_complete", Data: upload})
	return upload, nil
}
