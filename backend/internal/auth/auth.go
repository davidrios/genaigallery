package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"gorm.io/gorm"

	"genai-gallery-backend/internal/database"
	"genai-gallery-backend/internal/models"
)

var GlobalBearerToken string
var GlobalBasicAuthPassword string
var GlobalJWTSecret []byte

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

	var jwtConfig models.AppConfig
	err = db.Where("key = ?", "jwt_secret").First(&jwtConfig).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			bytes := make([]byte, 32)
			if _, randErr := rand.Read(bytes); randErr != nil {
				log.Fatalf("failed to generate JWT secret: %v", randErr)
			}
			jwtConfig = models.AppConfig{
				Key:   "jwt_secret",
				Value: hex.EncodeToString(bytes),
			}
			if insertErr := db.Create(&jwtConfig).Error; insertErr != nil {
				log.Fatalf("failed to save JWT secret to database: %v", insertErr)
			}
		} else {
			log.Fatalf("failed to fetch JWT secret from database: %v", err)
		}
	}

	var decodeErr error
	GlobalJWTSecret, decodeErr = hex.DecodeString(jwtConfig.Value)
	if decodeErr != nil {
		log.Fatalf("failed to decode JWT secret: %v", decodeErr)
	}

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

var jwtHeader = base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))

// GenerateImageToken creates a short-lived HS256 JWT for image URL access
func GenerateImageToken() string {
	exp := time.Now().Add(1 * time.Minute).Unix()
	payload := base64.RawURLEncoding.EncodeToString(fmt.Appendf(nil, `{"exp":%d}`, exp))

	data := jwtHeader + "." + payload
	mac := hmac.New(sha256.New, GlobalJWTSecret)
	mac.Write([]byte(data))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return data + "." + sig
}

// ValidateImageToken verifies the HS256 signature and expiry of an image JWT.
func ValidateImageToken(token string) bool {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return false
	}

	data := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, GlobalJWTSecret)
	mac.Write([]byte(data))
	expected := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(parts[2]), []byte(expected)) {
		return false
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return false
	}

	var payload struct {
		Exp int64 `json:"exp"`
	}
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return false
	}

	return time.Now().Unix() < payload.Exp
}
