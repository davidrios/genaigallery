package handlers

import (
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"genai-gallery-backend/internal/config"
	"genai-gallery-backend/internal/database"
	"genai-gallery-backend/internal/models"
	gallerysync "genai-gallery-backend/internal/sync"
)

func GetImageCore(id int) (*models.Image, error) {
	var image models.Image
	if err := database.GetDB().Preload("MetadataItems").First(&image, id).Error; err != nil {
		return nil, err
	}
	return &image, nil
}

func GetImage(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	image, err := GetImageCore(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Image not found"})
		return
	}
	c.JSON(http.StatusOK, image)
}

type Directory struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type BrowseResult struct {
	Directories []Directory    `json:"directories"`
	Images      []models.Image `json:"images"`
	Total       int            `json:"total"`
	Page        int            `json:"page"`
	Pages       int            `json:"pages"`
}

func BrowseCore(pathParam, q, sortOrder string, page, limit int) (*BrowseResult, error) {
	db := database.GetDB()
	gallerysync.CheckSync(db)

	pathParam = strings.TrimLeft(pathParam, "/")
	if strings.Contains(pathParam, "..") {
		return nil, fmt.Errorf("Invalid path")
	}

	fullPath := filepath.Join(config.ImagesDir, pathParam)
	info, err := os.Stat(fullPath)
	if err != nil || !info.IsDir() {
		return nil, fmt.Errorf("Directory not found")
	}

	var directories []Directory
	if q == "" {
		entries, err := os.ReadDir(fullPath)
		if err == nil {
			for _, entry := range entries {
				if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
					dirPath := entry.Name()
					if pathParam != "" {
						dirPath = filepath.Join(pathParam, entry.Name())
					}
					directories = append(directories, Directory{
						Name: entry.Name(),
						Path: dirPath,
					})
				}
			}
		}
	}

	query := db.Model(&models.Image{})

	if q != "" {
		if strings.Contains(q, ":") {
			parts := strings.SplitN(q, ":", 2)
			key, val := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
			query = query.Joins("JOIN image_metadata ON image_metadata.image_id = images.id").
				Where("image_metadata.key = ? AND image_metadata.value LIKE ?", key, "%"+val+"%")
		} else {
			var ids []uint
			ftsQuery := "SELECT image_id FROM search_index WHERE search_index MATCH ?"
			db.Raw(ftsQuery, "\""+q+"\"").Scan(&ids)
			if len(ids) > 0 {
				query = query.Where("id IN ?", ids)
			} else {
				query = query.Where("1 = 0")
			}
		}
	} else {
		if pathParam != "" {
			prefix := pathParam + string(os.PathSeparator)
			query = query.Where("path LIKE ?", prefix+"%")
		}
	}

	if sortOrder == "asc" {
		query = query.Order("created_at asc")
	} else {
		query = query.Order("created_at desc")
	}

	var allImages []models.Image
	query.Find(&allImages)

	var filteredImages []models.Image
	if q != "" {
		filteredImages = allImages
	} else {
		for _, img := range allImages {
			dir := filepath.Dir(img.Path)
			if dir == "." {
				dir = ""
			}
			if dir == pathParam {
				filteredImages = append(filteredImages, img)
			}
		}
	}

	total := len(filteredImages)
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	start := (page - 1) * limit
	end := start + limit
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	var paginated []models.Image
	if start < end {
		paginated = filteredImages[start:end]
	} else {
		paginated = make([]models.Image, 0)
	}
	if directories == nil {
		directories = make([]Directory, 0)
	}

	return &BrowseResult{
		Directories: directories,
		Images:      paginated,
		Total:       total,
		Page:        page,
		Pages:       totalPages,
	}, nil
}

func Browse(c *gin.Context) {
	pathParam := c.Query("path")
	q := c.Query("q")
	sortOrder := c.Query("sort")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if limit < 1 {
		limit = 50
	}

	result, err := BrowseCore(pathParam, q, sortOrder, page, limit)
	if err != nil {
		if err.Error() == "Invalid path" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, result)
}

func UploadCore(files []*multipart.FileHeader, prefix string) ([]*models.Image, error) {
	prefix = strings.Trim(prefix, "/")

	if strings.Contains(prefix, "..") {
		return nil, errors.New("invalid prefix path")
	}

	dirName := filepath.Dir(prefix)
	baseName := filepath.Base(prefix)
	if baseName == "." {
		baseName = ""
	}

	fullDir := filepath.Join(config.ImagesDir, dirName)
	err := os.MkdirAll(fullDir, 0755)
	if err != nil {
		return nil, err
	}

	db := database.GetDB()
	var createdImages []*models.Image

	for _, file := range files {
		timestamp := time.Now().Format("20060102150405")
		filename := fmt.Sprintf("%s_%s%s", baseName, timestamp, filepath.Ext(file.Filename))
		savePath := filepath.Join(fullDir, filename)

		src, err := file.Open()
		if err != nil {
			log.Printf("Error opening uploaded file: %v", err)
			continue
		}

		dst, err := os.OpenFile(savePath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
		if err != nil {
			log.Printf("Error creating image file: %v", err)
			src.Close()
			continue
		}

		_, err = io.Copy(dst, src)
		src.Close()
		dst.Close()

		if err != nil {
			log.Printf("Error copying image file contents: %v", err)
			continue
		}

		img, err := gallerysync.AddImage(db, savePath, time.Now(), false)
		if err != nil {
			log.Printf("Error adding image: %v", err)
			continue
		}

		createdImages = append(createdImages, img)
	}

	return createdImages, nil
}

func Upload(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	files := form.File["files"]
	prefix := c.PostForm("filename_prefix")

	createdImages, err := UploadCore(files, prefix)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, createdImages)
}
