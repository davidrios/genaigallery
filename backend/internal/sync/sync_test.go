package sync

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"genai-gallery-backend/internal/config"
	"genai-gallery-backend/internal/database"
	"genai-gallery-backend/internal/models"
)

func TestFindModifiedMedia(t *testing.T) {
	rootDir := "fixtures"

	zeroTime := time.Time{}

	var discoveredFiles []string
	err := findModifiedMedia(rootDir, zeroTime, func(path string, d os.DirEntry) {
		discoveredFiles = append(discoveredFiles, path)
	})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(discoveredFiles) < 5 {
		t.Errorf("Expected at least 5 media files, got %d", len(discoveredFiles))
	}

	for _, file := range discoveredFiles {
		ext := strings.ToLower(filepath.Ext(file))
		if ext != ".png" && ext != ".jpg" && ext != ".jpeg" && ext != ".webp" && ext != ".mp4" && ext != ".mkv" && ext != ".mov" && ext != ".webm" {
			t.Errorf("Found unexpected file type: %s", file)
		}
	}

	futureTime := time.Now().Add(24 * time.Hour)
	var futureDiscoveredFiles []string
	err = findModifiedMedia(rootDir, futureTime, func(path string, d os.DirEntry) {
		futureDiscoveredFiles = append(futureDiscoveredFiles, path)
	})

	if err != nil {
		t.Fatalf("Expected no error on future time walk, got %v", err)
	}

	if len(futureDiscoveredFiles) > 0 {
		t.Errorf("Expected 0 files modified in the future, got %d", len(futureDiscoveredFiles))
	}
}

func TestPerformSync(t *testing.T) {
	originalImagesDir := config.ImagesDir
	originalDBPath := config.DBPath

	defer func() {
		config.ImagesDir = originalImagesDir
		config.DBPath = originalDBPath
	}()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Could not determine cwd: %v", err)
	}

	config.ImagesDir = filepath.Join(cwd, "fixtures")
	config.DBPath = filepath.Join(cwd, "fixtures", "test.db")

	err = os.Remove(config.DBPath)
	if err != nil && !os.IsNotExist(err) {
		t.Fatalf("unable to clear fixture db")
	}

	database.InitDB(config.DBPath)
	db := database.GetDB()

	performSync(db)

	var images []models.Image
	db.Find(&images)

	if len(images) < 5 {
		t.Fatalf("Expected at least 5 media files synced on fresh db, got %d", len(images))
	}

	// Make sure paths are relative to ImagesDir
	for _, img := range images {
		if strings.HasPrefix(img.Path, config.ImagesDir) {
			t.Errorf("Image path %s should be relative to ImagesDir, but is absolute", img.Path)
		}
	}

	// Verify metadata extraction roughly worked (the PNGs have metadata)
	var metaCount int64
	db.Raw("SELECT COUNT(1) FROM search_index").Scan(&metaCount)
	if metaCount == 0 {
		t.Errorf("Expected metadata to be extracted from test images, but got 0 meta rows")
	}

	// 2. Secondary Sync (No changes)
	performSync(db)

	var secondImageSet []models.Image
	db.Find(&secondImageSet)

	if len(secondImageSet) != len(images) {
		t.Errorf("Secondary sync changed image count: expected %d, got %d", len(images), len(secondImageSet))
	}

	newTime := time.Now()
	err = os.Chtimes(filepath.Join(config.ImagesDir, "ComfyUI_00001_.png"), newTime, newTime)
	if err != nil {
		t.Fatalf("unable to change file time")
	}

	var thirdImageSet []models.Image
	db.Find(&thirdImageSet)
}
