package entity

type Settings struct {
	ID                    uint   `gorm:"primaryKey" json:"id"`
	UploadDirectory       string `gorm:"not null" json:"upload_directory"`
	AutoOrganize          bool   `gorm:"not null;default:true" json:"auto_organize"`
	MaxUploadSize         *int64 `json:"max_upload_size"`
	MaxConcurrentUploads  int    `gorm:"not null;default:3" json:"max_concurrent_uploads"`
	SessionTimeoutMinutes int    `gorm:"not null;default:30" json:"session_timeout_minutes"`
	AutoStartService      bool   `gorm:"not null;default:false" json:"auto_start_service"`
}
