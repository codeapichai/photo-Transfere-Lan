package entity

import "time"

type ActivityLog struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	EventType    string    `gorm:"not null" json:"event_type"`
	Actor        string    `json:"actor"`
	Message      string    `gorm:"not null" json:"message"`
	MetadataJSON string    `json:"metadata_json"`
	CreatedAt    time.Time `json:"created_at"`
}
