package sync

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gorm.io/gorm"

	"genai-gallery-backend/internal/config"
	"genai-gallery-backend/internal/database"
	"genai-gallery-backend/internal/models"
)

func TestFindModifiedMedia(t *testing.T) {
	rootDir := "../../testdata/fixtures"

	zeroTime := time.Time{}

	var discoveredFiles []string
	err := findModifiedMedia(rootDir, zeroTime, func(path string, modtime time.Time, syncFile bool) {
		if syncFile {
			discoveredFiles = append(discoveredFiles, path)
		}
	})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(discoveredFiles) < 5 {
		t.Errorf("Expected at least 5 media files, got %d", len(discoveredFiles))
	}

	for _, file := range discoveredFiles {
		ext := strings.ToLower(filepath.Ext(file))
		if ext != ".png" && ext != ".jpg" && ext != ".jpeg" && ext != ".webp" && ext != ".mp4" && ext != ".mov" && ext != ".webm" {
			t.Errorf("Found unexpected file type: %s", file)
		}
	}

	futureTime := time.Now().Add(24 * time.Hour)
	var futureDiscoveredFiles []string
	err = findModifiedMedia(rootDir, futureTime, func(path string, modtime time.Time, syncFile bool) {
		if syncFile {
			futureDiscoveredFiles = append(futureDiscoveredFiles, path)
		}
	})

	if err != nil {
		t.Fatalf("Expected no error on future time walk, got %v", err)
	}

	if len(futureDiscoveredFiles) > 0 {
		t.Errorf("Expected 0 files modified in the future, got %d", len(futureDiscoveredFiles))
	}
}

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	database.InitDB("file::memory:")

	return database.GetDB()
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

	config.ImagesDir = filepath.Join(cwd, "..", "..", "testdata", "fixtures")
	config.DBPath = filepath.Join(cwd, "..", "..", "testdata", "fixtures", "test.db")

	err = os.Remove(config.DBPath)
	if err != nil && !os.IsNotExist(err) {
		t.Fatalf("unable to clear fixture db")
	}

	database.InitDB(config.DBPath)
	db := database.GetDB()

	PerformSync(db)

	var images []models.Image
	db.Find(&images)

	if len(images) < 5 {
		t.Fatalf("Expected at least 5 media files synced on fresh db, got %d", len(images))
	}

	for _, img := range images {
		// Make sure paths are relative to ImagesDir
		if strings.HasPrefix(img.Path, config.ImagesDir) {
			t.Errorf("Image path %s should be relative to ImagesDir, but is absolute", img.Path)
		}

		if strings.Contains(img.Path, "\\") {
			t.Errorf("Image path %s must use forward slashes, but contains backslashes", img.Path)
		}

		relPath := filepath.Join(filepath.FromSlash(img.Path), img.Name)

		// Verify correct CreatedAt
		info, err := os.Stat(filepath.Join(config.ImagesDir, relPath))
		if err != nil {
			t.Fatalf("Couldnt stat image file: %v", err)
		}

		if !img.CreatedAt.Equal(info.ModTime()) {
			t.Errorf("Image %s should have CreatedAt %s but has %s", relPath, info.ModTime(), img.CreatedAt)
		}
	}

	// Verify metadata extraction roughly worked (all files have metadata)
	var metaCount int64
	db.Raw("select count(1) from (SELECT image_id, COUNT(1) FROM image_metadata GROUP BY image_id) t1").Scan(&metaCount)
	if metaCount != 5 {
		t.Errorf("Expected metadata to be extracted from %v test images, but got %v meta rows", 5, metaCount)
	}

	// second sync (no changes)
	PerformSync(db)

	var secondImageSet []models.Image
	db.Find(&secondImageSet)

	if len(secondImageSet) != len(images) {
		t.Errorf("Secondary sync changed image count: expected %d, got %d", len(images), len(secondImageSet))
	}

	info, err := os.Stat(config.DBPath)
	if err != nil {
		t.Fatal("Couldnt stat database file")
	}

	t.Log(info.ModTime())

	newTime := info.ModTime().Add(time.Second * 10)
	err = os.Chtimes(filepath.Join(config.ImagesDir, "ComfyUI_00001_.png"), newTime, newTime)
	if err != nil {
		t.Fatalf("unable to change file time")
	}

	// third sync, 1 updated file
	PerformSync(db)

	var thirdImageSet []models.Image
	db.Find(&thirdImageSet)

	if len(thirdImageSet) != len(images) {
		t.Fatalf("Third sync changed total image count: expected %d, got %d", len(images), len(thirdImageSet))
	}

	// Compare IDs between initial set and third set
	initialIDs := make(map[uint]bool)
	for _, img := range images {
		initialIDs[img.ID] = true
	}

	matchCount := 0
	diffCount := 0
	for _, img := range thirdImageSet {
		if initialIDs[img.ID] {
			matchCount++
		} else {
			diffCount++
		}
	}

	if matchCount != len(images)-1 {
		t.Errorf("Expected %d matching IDs, got %d", len(images)-1, matchCount)
	}
	if diffCount != 1 {
		t.Errorf("Expected exactly 1 new/different ID, got %d", diffCount)
	}
}

func TestUpdateFTS(t *testing.T) {
	db := setupTestDB(t)

	// Stub an image
	img := models.Image{
		Path: "",
		Name: "test_prefix.png",
	}
	db.Create(&img)

	meta := []models.ImageMetadata{
		{Key: "meta10", Value: "apple"},
		{Key: "meta123", Value: "banana"},
		{Key: "meta999", Value: "cherry"},
		{Key: "other", Value: "dog"},
	}

	updateFTS(db, &img, &meta)

	// Verify the database state using Raw SQL
	type SearchRow struct {
		ImageID uint   `gorm:"column:image_id"`
		Content string `gorm:"column:content"`
	}

	var rows []SearchRow
	err := db.Raw("SELECT image_id, content FROM search_index WHERE image_id = ?", img.ID).Scan(&rows).Error
	if err != nil {
		t.Fatalf("Failed to query search_index: %v", err)
	}

	if len(rows) != 1 {
		t.Fatalf("Expected 1 row in search_index, got %d", len(rows))
	}

	row := rows[0]

	if !strings.Contains(row.Content, "apple") || !strings.Contains(row.Content, "banana") || !strings.Contains(row.Content, "cherry") {
		t.Errorf("Content contains unexpected data: %s", row.Content)
	}
}
