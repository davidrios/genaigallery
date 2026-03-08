package sync

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/charlievieth/fastwalk"
	"gorm.io/gorm"

	"genai-gallery-backend/internal/config"
	"genai-gallery-backend/internal/metadata"
	"genai-gallery-backend/internal/models"
)

var (
	syncDone bool
	syncLock sync.Mutex
)

func CheckSync(db *gorm.DB) {
	if syncDone {
		return
	}

	if !syncLock.TryLock() {
		return
	}
	defer syncLock.Unlock()

	performSync(db)
	syncDone = true
}

func performSync(db *gorm.DB) {
	log.Println("Starting sync...")

	info, err := os.Stat(config.DBPath)
	modTime := time.Time{}
	if err != nil {
		modTime = info.ModTime()
	}

	var syncedCount uint64

	err = findModifiedMedia(config.ImagesDir, modTime, func(path string, d os.DirEntry) {
		relPath, err := filepath.Rel(config.ImagesDir, path)
		if err != nil {
			return
		}

		var image models.Image
		query := db.Model(&models.Image{}).Where("path = ?", relPath)
		if err := query.First(&image).Error; err != nil {
			db.Exec("DELETE FROM search_index WHERE image_id = ?", image.ID)
			db.Delete(image)
		}

		newImg := models.Image{
			Path:      relPath,
			CreatedAt: info.ModTime(),
		}

		if err := db.Create(&newImg).Error; err == nil {
			extractAndSaveMetadata(db, &newImg, path)
			atomic.AddUint64(&syncedCount, 1)
		}
	})

	if err != nil {
		log.Println("Error syncing")
		return
	}

	log.Printf("Synced %d files.\n", atomic.LoadUint64(&syncedCount))
}

func extractAndSaveMetadata(db *gorm.DB, img *models.Image, fullPath string) {
	if !strings.HasSuffix(strings.ToLower(fullPath), ".png") {
		updateFTS(db, img, nil)
		return
	}

	items, err := metadata.ExtractMetadata(fullPath)
	if err != nil {
		updateFTS(db, img, nil)
		return
	}

	var metaRows []models.ImageMetadata

	for _, item := range items {
		metaRows = append(metaRows, models.ImageMetadata{
			ImageID: img.ID,
			Key:     item.Key,
			Value:   item.Value,
		})
	}

	// if len(metaRows) > 0 {
	// 	db.Create(&metaRows)
	// }

	updateFTS(db, img, &metaRows)
}

func updateFTS(db *gorm.DB, img *models.Image, metaData *[]models.ImageMetadata) {
	db.Exec("DELETE FROM search_index WHERE image_id = ?", img.ID)

	fullContent := img.Path + " " + img.Prompt
	db.Exec("INSERT INTO search_index (image_id, prefix, content) VALUES (?, '', ?)", img.ID, fullContent)
	if metaData != nil {
		for i := range *metaData {
			metaItem := &(*metaData)[i]
			db.Exec("INSERT INTO search_index (image_id, prefix, content) VALUES (?, ?, ?)", img.ID, metaItem.Key, metaItem.Value)
		}
	}
}

func findModifiedMedia(rootDir string, dbModTime time.Time, processFunc func(string, os.DirEntry)) error {
	err := fastwalk.Walk(nil, rootDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if d.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(d.Name()))
		if ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".webp" || ext == ".mp4" || ext == ".mov" || ext == ".mkv" || ext == ".webm" {
			info, err := d.Info()
			if err == nil {
				if info.ModTime().After(dbModTime) {
					processFunc(path, d)
				}
			}
		}
		return nil
	})

	return err
}
