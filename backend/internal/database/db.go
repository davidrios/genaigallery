package database

import (
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"genai-gallery-backend/internal/models"
)

var DB *gorm.DB

func InitDB(dbPath string) {
	var err error
	DB, err = gorm.Open(sqlite.Open(dbPath+"?_journal_mode=WAL&_busy_timeout=5000"), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database:", err)
	}

	// AutoMigrate
	err = DB.AutoMigrate(&models.Image{}, &models.ImageMetadata{})
	if err != nil {
		log.Fatal("failed to migrate database:", err)
	}

	// Create FTS5 table
	// Note: We use image_id as UNINDEXED column to reference the Image.ID (Integer)
	err = DB.Exec("CREATE VIRTUAL TABLE IF NOT EXISTS search_index USING fts5(image_id UNINDEXED, prefix UNINDEXED, content, tokenize=\"trigram\")").Error
	if err != nil {
		log.Fatal("failed to create fts table:", err)
	}
}

func GetDB() *gorm.DB {
	return DB
}
