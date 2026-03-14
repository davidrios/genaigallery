package config

import (
	"os"
	"path/filepath"
)

var (
	ImagesDir string
	DBPath    string
	Port      string
)

const StaticImagesRoot = "/images"

func InitConfig() {
	ImagesDir = os.Getenv("IMAGES_DIR")
	if ImagesDir == "" {
		cwd, _ := os.Getwd()
		ImagesDir = cwd
	}

	DBPath = os.Getenv("DB_PATH")
	if DBPath == "" {
		DBPath = filepath.Join(ImagesDir, "genaigallery.db")
	}

	Port = os.Getenv("PORT")
	if Port == "" {
		Port = "8000"
	}
}
