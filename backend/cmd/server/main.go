package main

import (
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"genai-gallery-backend/internal/config"
	"genai-gallery-backend/internal/database"
	"genai-gallery-backend/internal/handlers"
)

func main() {
	config.InitConfig()
	database.InitDB(config.DBPath)

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:" + config.Port},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type"},
		AllowCredentials: true,
	}))

	r.Static("/images", config.ImagesDir)
	// r.Static("/assets", ...) // Frontend assets if needed

	api := r.Group("/api")
	{
		// api.GET("/images", handlers.ListImages)
		api.GET("/image/:id", handlers.GetImage)
		api.GET("/browse", handlers.Browse)
		// api.GET("/search", handlers.ListImages) // Search is basically List with q
		api.POST("/upload", handlers.Upload)
	}

	log.Printf("Server starting on port %s", config.Port)
	if err := r.Run(":" + config.Port); err != nil {
		log.Fatal(err)
	}
}
