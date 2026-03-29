package sync

import (
	"errors"
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

	PerformSync(db)
	syncDone = true
}

func AddImage(db *gorm.DB, origPath string, modtime time.Time, replace bool) (*models.Image, error) {
	relPath, err := filepath.Rel(config.ImagesDir, origPath)
	if err != nil {
		return nil, err
	}

	path := filepath.Dir(relPath)
	if path == "." {
		path = ""
	}
	path = filepath.ToSlash(path)
	name := filepath.Base(relPath)

	var image models.Image
	err = db.Model(&models.Image{}).Where("path = ? and name = ?", path, name).First(&image).Error

	if err == nil {
		if replace {
			db.Exec("DELETE FROM search_index WHERE image_id = ?", image.ID)
			db.Exec("DELETE FROM image_metadata WHERE image_id = ?", image.ID)
			db.Delete(&image)
		} else {
			return nil, errors.New("image already exists in DB")
		}
	} else if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	newImg := models.Image{
		Path:      path,
		Name:      name,
		CreatedAt: modtime,
	}

	err = db.Create(&newImg).Error
	if err != nil {
		return nil, err
	}

	extractAndSaveMetadata(db, &newImg, origPath)
	return &newImg, nil
}

func PerformSync(db *gorm.DB) {
	log.Println("Starting sync, this may take a while...")

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
	var scannedCount uint64

	err = findModifiedMedia(config.ImagesDir, modTime, func(path string, modtime time.Time, syncFile bool) {
		partialScanned := atomic.AddUint64(&scannedCount, 1)
		if partialScanned%5000 == 0 {
			log.Printf("Scanned %d files, continuing...\n", partialScanned)
		}

		if !syncFile {
			return
		}

		_, err := AddImage(db, path, modtime, true)
		if err != nil {
			log.Printf("Error adding image %s: %v", path, err)
			return
		}

		atomic.AddUint64(&syncedCount, 1)
	})

	if err != nil {
		log.Println("Error syncing")
		return
	}

	log.Printf("Scanned %d files, synced %d.\n", atomic.LoadUint64(&scannedCount), atomic.LoadUint64(&syncedCount))
}

func extractAndSaveMetadata(db *gorm.DB, img *models.Image, fullPath string) {
	pathLower := strings.ToLower(fullPath)
	if !(strings.HasSuffix(pathLower, ".png") || strings.HasSuffix(pathLower, ".mp4")) {
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

	fullContent := []string{img.Path, img.Name, img.CreatedAt.String()}

	if metaData != nil && len(*metaData) > 0 {
		for i := range *metaData {
			metaItem := &(*metaData)[i]

			// Extract prefix by stripping trailing digits
			prefix := metaItem.Key
			for len(prefix) > 0 && prefix[len(prefix)-1] >= '0' && prefix[len(prefix)-1] <= '9' {
				prefix = prefix[:len(prefix)-1]
			}

			fullContent = append(fullContent, prefix+":"+metaItem.Value)
		}
	}

	db.Exec("INSERT INTO search_index (image_id, content) VALUES (?, ?)", img.ID, strings.Join(fullContent, " "))
}

func findModifiedMedia(rootDir string, dbModTime time.Time, processFunc func(string, time.Time, bool)) error {
	err := fastwalk.Walk(nil, rootDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if d.IsDir() {
			if d.Name() == ".video_preview" {
				return filepath.SkipDir
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(d.Name()))
		if !(ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".webp" || ext == ".mp4" || ext == ".mov" || ext == ".webm") {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}

		processFunc(path, info.ModTime(), info.ModTime().After(dbModTime))

		return nil
	})

	return err
}
