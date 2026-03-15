package config

import (
	"flag"
	"os"
	"path/filepath"
)

var (
	ImagesDir   string
	DBPath      string
	Port        string
	RequireAuth bool
)

const StaticImagesRoot = "/images"

func InitConfig() {
	cwd, _ := os.Getwd()

	defaultImagesDir := os.Getenv("IMAGES_DIR")
	if defaultImagesDir == "" {
		defaultImagesDir = cwd
	}
	flag.StringVar(&ImagesDir, "images-dir", defaultImagesDir, "Directory containing the images (or set IMAGES_DIR env var)")

	defaultDbPath := os.Getenv("DB_PATH")
	var dbPathArg string
	flag.StringVar(&dbPathArg, "db-path", defaultDbPath, "Path to the SQLite database file (or set DB_PATH env var)")

	defaultPort := os.Getenv("PORT")
	flag.StringVar(&Port, "port", defaultPort, "Port to run the server on (or set PORT env var)")

	defaultRequireAuth := os.Getenv("REQUIRE_AUTH") == "true"
	flag.BoolVar(&RequireAuth, "require-auth", defaultRequireAuth, "Require authentication for all networks (or set REQUIRE_AUTH=true env var)")

	flag.Parse()

	DBPath = dbPathArg
	if DBPath == "" {
		DBPath = filepath.Join(ImagesDir, "genaigallery.db")
	}
}
