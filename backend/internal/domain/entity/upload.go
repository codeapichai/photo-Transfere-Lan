package entity

import "time"

type UploadStatus string

const (
	UploadPending   UploadStatus = "pending"
	UploadUploading UploadStatus = "uploading"
	UploadSuccess   UploadStatus = "success"
	UploadFailed    UploadStatus = "failed"
	UploadCorrupted UploadStatus = "corrupted"
	UploadDuplicate UploadStatus = "duplicate"
)

type Upload struct {
	ID               string       `gorm:"primaryKey;size:36" json:"id"`
	Filename         string       `gorm:"not null" json:"filename"`
	OriginalFilename string       `gorm:"not null" json:"original_filename"`
	Filesize         int64        `gorm:"not null" json:"filesize"`
	SHA256           string       `gorm:"index" json:"sha256"`
	UploadTime       time.Time    `json:"upload_time"`
	DeviceName       string       `json:"device_name"`
	Status           UploadStatus `gorm:"size:32;not null" json:"status"`
	StoragePath      string       `gorm:"not null" json:"storage_path"`
	ReceivedBytes    int64        `gorm:"not null;default:0" json:"received_bytes"`
}

type DuplicatePolicy string

const (
	DuplicateSkip      DuplicatePolicy = "skip"
	DuplicateOverwrite DuplicatePolicy = "overwrite"
	DuplicateKeepBoth  DuplicatePolicy = "keep_both"
)
