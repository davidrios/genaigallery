package database

import (
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"genai-gallery-backend/internal/models"
)

var DB *gorm.DB

func InitDB(dbPath string) {
	var err error
	DB, err = gorm.Open(sqlite.Open(dbPath+"?_journal_mode=WAL&_busy_timeout=5000"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info).LogMode(logger.Warn).LogMode(logger.Error),
	})
	if err != nil {
		log.Fatal("failed to connect database:", err)
	}

	// Reconfigure logger specifically to ignore 'record not found'
	DB.Logger = logger.Default.LogMode(logger.Warn)
	newLogger := logger.New(
		log.New(log.Writer(), "\r\n", log.LstdFlags),
		logger.Config{
			IgnoreRecordNotFoundError: true,
		},
	)
	DB, err = gorm.Open(sqlite.Open(dbPath+"?_journal_mode=WAL&_busy_timeout=5000"), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		log.Fatal("failed to connect database:", err)
	}

	// AutoMigrate
	err = DB.AutoMigrate(&models.Image{}, &models.ImageMetadata{}, &models.AppConfig{})
	if err != nil {
		log.Fatal("failed to migrate database:", err)
	}

	// Create FTS5 table
	// Note: We use image_id as UNINDEXED column to reference the Image.ID (Integer)
	err = DB.Exec("CREATE VIRTUAL TABLE IF NOT EXISTS search_index USING fts5(image_id UNINDEXED, content, tokenize=\"trigram\")").Error
	if err != nil {
		log.Fatal("failed to create fts table:", err)
	}
}

func GetDB() *gorm.DB {
	return DB
}
