package database

import (
	"phototransferlan/backend/internal/domain/entity"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func Open(path string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&entity.User{}, &entity.Upload{}, &entity.Settings{}, &entity.ActivityLog{}, &entity.Session{}, &entity.TemporaryToken{}); err != nil {
		return nil, err
	}
	return db, nil
}
