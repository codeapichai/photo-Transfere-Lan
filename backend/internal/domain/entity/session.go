package entity

import "time"

type Session struct {
	ID        string    `gorm:"primaryKey;size:64" json:"id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	CSRFToken string    `gorm:"size:64;not null" json:"csrf_token"`
	ExpiresAt time.Time `gorm:"not null;index" json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

type TemporaryToken struct {
	ID        string    `gorm:"primaryKey;size:64" json:"id"`
	TokenHash string    `gorm:"size:64;not null;uniqueIndex" json:"-"`
	ExpiresAt time.Time `gorm:"not null;index" json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}
