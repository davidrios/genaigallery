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
	if err != nil {
		log.Fatalln("Couldnt stat database file")
	}

	modTime := info.ModTime()
	var count int64
	db.Model(&models.Image{}).Count(&count)
	if count == 0 {
		modTime = time.Time{}
	}

	var syncedCount uint64

	err = findModifiedMedia(config.ImagesDir, modTime, func(path string, d os.DirEntry) {
		relPath, err := filepath.Rel(config.ImagesDir, path)
		if err != nil {
			return
		}

		var image models.Image
		err = db.Model(&models.Image{}).Where("path = ?", relPath).First(&image).Error

		if err == nil {
			db.Exec("DELETE FROM search_index WHERE image_id = ?", image.ID)
			db.Exec("DELETE FROM image_metadata WHERE image_id = ?", image.ID)
			db.Delete(&image)
		} else if err != gorm.ErrRecordNotFound {
			log.Printf("DB error querying image %s: %v", relPath, err)
			return
		}

		newImg := models.Image{
			Path:      relPath,
			CreatedAt: info.ModTime(),
		}

		if err := db.Create(&newImg).Error; err == nil {
			extractAndSaveMetadata(db, &newImg, path)
			atomic.AddUint64(&syncedCount, 1)
		} else {
			log.Printf("Failed to create image record for %s: %v", relPath, err)
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

	if len(metaRows) > 0 {
		db.Create(&metaRows)
	}

	if len(metaRows) > 0 {
		updateFTS(db, img, &metaRows)
	} else {
		updateFTS(db, img, nil)
	}
}

func updateFTS(db *gorm.DB, img *models.Image, metaData *[]models.ImageMetadata) {
	db.Exec("DELETE FROM search_index WHERE image_id = ?", img.ID)

	fullContent := img.Path + " " + img.Prompt
	db.Exec("INSERT INTO search_index (image_id, prefix, content) VALUES (?, '', ?)", img.ID, fullContent)
	if metaData != nil && len(*metaData) > 0 {
		grouped := make(map[string]string)

		for i := range *metaData {
			metaItem := &(*metaData)[i]

			// Extract prefix by stripping trailing digits
			prefix := metaItem.Key
			for len(prefix) > 0 && prefix[len(prefix)-1] >= '0' && prefix[len(prefix)-1] <= '9' {
				prefix = prefix[:len(prefix)-1]
			}

			if existing, ok := grouped[prefix]; ok {
				grouped[prefix] = existing + " " + metaItem.Value
			} else {
				grouped[prefix] = metaItem.Value
			}
		}

		for prefix, content := range grouped {
			db.Exec("INSERT INTO search_index (image_id, prefix, content) VALUES (?, ?, ?)", img.ID, prefix, content)
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
