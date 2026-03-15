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
var GlobalBasicAuthPassword string

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

	var basicAuthConfig models.AppConfig
	err = db.Where("key = ?", "basic_auth_password").First(&basicAuthConfig).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Generate a new 16-byte secure random hex password
			bytes := make([]byte, 16)
			if _, randErr := rand.Read(bytes); randErr != nil {
				log.Fatalf("failed to generate secure crypto random basic auth password: %v", randErr)
			}
			newPassword := hex.EncodeToString(bytes)

			basicAuthConfig = models.AppConfig{
				Key:   "basic_auth_password",
				Value: newPassword,
			}

			if insertErr := db.Create(&basicAuthConfig).Error; insertErr != nil {
				log.Fatalf("failed to save generated basic auth password to database: %v", insertErr)
			}
		} else {
			log.Fatalf("failed to fetch basic auth config from database: %v", err)
		}
	}

	GlobalBasicAuthPassword = basicAuthConfig.Value

	fmt.Println()
	fmt.Println("==========================================================================")
	fmt.Println(" GenAI Gallery - Remote Access Configuration")
	fmt.Println("==========================================================================")
	fmt.Println(" To connect from a remote network or browser, use the following:")
	fmt.Println()
	fmt.Printf(" [API/Client] Bearer %s\n", GlobalBearerToken)
	fmt.Printf(" [Web UI]     Username: (leave blank)\n")
	fmt.Printf("              Password: %s\n", GlobalBasicAuthPassword)
	fmt.Println("==========================================================================")
	fmt.Println()
}
