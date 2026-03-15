package main

import (
	"archive/zip"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"genai-gallery-backend/internal/config"
	"genai-gallery-backend/internal/database"
	"genai-gallery-backend/internal/handlers"
)

func serveStaticZip(r *gin.Engine) {
	exePath, err := os.Executable()
	if err != nil {
		return
	}

	exe := filepath.Base(exePath)
	if strings.Index(exe, "genaigallery") != 0 {
		return
	}

	file, err := os.Open(exePath)
	if err != nil {
		log.Fatalln("Could not read self executable")
		return
	}

	info, err := file.Stat()
	if err != nil {
		log.Fatalln("Could not stat self executable")
		return
	}

	reader, err := zip.NewReader(file, info.Size())
	if err != nil {
		log.Fatalf("error reading self zip: %v\n", err)
		return
	}

	rootFs, err := fs.Sub(reader, ".")
	if err != nil {
		log.Fatalf("error creating FS for self zip: %v\n", err)
		return
	}
	rootHttpFs := http.FS(rootFs)

	r.StaticFileFS("/favicon.ico", "favicon.ico", rootHttpFs)

	f, err := rootFs.Open("index.html")
	if err != nil {
		log.Fatalf("index.html not found in zip")
		return
	}
	defer f.Close()

	indexContent, err := io.ReadAll(f)
	if err != nil {
		log.Fatalf("error reading index.html")
		return
	}

	r.GET("/", func(c *gin.Context) {
		c.Data(200, "text/html; charset=utf-8", indexContent)
	})

	assetsFs, err := fs.Sub(reader, "assets")
	if err != nil {
		log.Fatalf("error creating FS for self zip: %v\n", err)
		return
	}

	r.StaticFS("/assets", http.FS(assetsFs))

	log.Println("Serving static files from self zip")
}

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

	r.Static(config.StaticImagesRoot, config.ImagesDir)

	api := r.Group("/api")
	{
		api.GET("/image/:id", handlers.GetImage)
		api.GET("/browse", handlers.Browse)
		api.POST("/upload", handlers.Upload)
	}

	serveStaticZip(r)

	log.Printf("Server starting on port %s", config.Port)
	if err := r.Run(":" + config.Port); err != nil {
		log.Fatal(err)
	}
}
