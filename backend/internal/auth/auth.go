package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"

	"gorm.io/gorm"

	"genai-gallery-backend/internal/database"
	"genai-gallery-backend/internal/models"
)

var GlobalBearerToken string

func InitAuth() {
	db := database.GetDB()
	var config models.AppConfig

	// Wrap search in db transaction for atomicity
	err := db.Where("key = ?", "bearer_token").First(&config).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Generate a new 32-byte secure random hex token
			bytes := make([]byte, 32)
			if _, randErr := rand.Read(bytes); randErr != nil {
				log.Fatalf("failed to generate secure crypto random token: %v", randErr)
			}
			newToken := hex.EncodeToString(bytes)

			config = models.AppConfig{
				Key:   "bearer_token",
				Value: newToken,
			}

			if insertErr := db.Create(&config).Error; insertErr != nil {
				log.Fatalf("failed to save generated bearer token to database: %v", insertErr)
			}
		} else {
			log.Fatalf("failed to fetch auth config from database: %v", err)
		}
	}

	GlobalBearerToken = config.Value

	fmt.Println()
	fmt.Println("==========================================================================")
	fmt.Println(" GenAI Gallery - Remote Access Configuration")
	fmt.Println("==========================================================================")
	fmt.Println(" To connect the frontend from a remote network, use this token:")
	fmt.Printf(" Bearer %s\n", GlobalBearerToken)
	fmt.Println("==========================================================================")
	fmt.Println()
}
