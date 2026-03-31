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
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"genai-gallery-backend/internal/auth"
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

// appendImageToken appends a short-lived JWT as a query parameter only when
// the request was authenticated via Bearer token.
func appendImageToken(c *gin.Context, path string) string {
	if !c.GetBool("bearer_authenticated") {
		return path
	}
	return path + "?token=" + auth.GenerateImageToken()
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

	basePath := image.Path

	var fullPath string
	if image.Path != "" {
		fullPath = config.StaticImagesRoot + "/" + image.Path + "/" + image.Name
	} else {
		fullPath = config.StaticImagesRoot + "/" + image.Name
	}
	image.Path = appendImageToken(c, fullPath)

	c.JSON(http.StatusOK, struct {
		*models.Image
		BasePath string `json:"base_path"`
	}{
		Image:    image,
		BasePath: basePath,
	})
}

type Directory struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type BrowseResultImage struct {
	*models.Image
	VideoPreview string `json:"video_preview"`
}

type BrowseResult struct {
	Directories []Directory         `json:"directories"`
	Images      []BrowseResultImage `json:"images"`
	Total       int64               `json:"total"`
	Page        int                 `json:"page"`
	Pages       int                 `json:"pages"`
}

func toFTSQuery(origStr string) string {
	if strings.Contains(origStr, "\"") {
		return origStr
	}

	ret := []string{}

	for word := range strings.SplitSeq(origStr, " ") {
		if word == "" {
			continue
		}

		if word == "OR" || word == "AND" {
			ret = append(ret, word)
		} else {
			ret = append(ret, fmt.Sprintf("\"%s\"", word))
		}
	}

	return strings.Join(ret, " ")
}

func generateVideoPreview(image *models.Image) string {
	if config.FfmpegPath == "" {
		return ""
	}

	var dir string
	if image.Path != "" {
		dir = filepath.Join(config.ImagesDir, image.Path)
	} else {
		dir = config.ImagesDir
	}

	previewDir := filepath.Join(dir, ".video_preview")
	previewName := image.Name + "__preview.jpg"
	previewFullPath := filepath.Join(previewDir, previewName)

	previewStaticBase := config.StaticImagesRoot
	if image.Path != "" {
		previewStaticBase += "/" + image.Path
	}
	previewStaticBase += "/.video_preview/" + previewName

	if _, err := os.Stat(previewFullPath); err == nil {
		return previewStaticBase
	}

	if err := os.MkdirAll(previewDir, 0755); err != nil {
		log.Printf("failed to create preview dir for %s: %v", image.Name, err)
		return ""
	}

	videoFullPath := filepath.Join(dir, image.Name)
	cmd := exec.Command(config.FfmpegPath, "-i", videoFullPath, "-vframes", "1", "-q:v", "2", previewFullPath)
	if err := cmd.Run(); err != nil {
		log.Printf("ffmpeg preview generation failed for %s: %v", image.Name, err)
		return ""
	}

	return previewStaticBase
}

func BrowseCore(pathParam string, q string, inPath bool, sortOrder string, page, limit int) (*BrowseResult, error) {
	db := database.GetDB()
	gallerysync.CheckSync(db)

	pathParam = strings.TrimLeft(pathParam, "/")
	if strings.Contains(pathParam, ".") {
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
						Path: filepath.ToSlash(dirPath),
					})
				}
			}
		}
	}

	query := db.Model(&models.Image{})

	if sortOrder != "asc" {
		sortOrder = "desc"
	}

	if q != "" {
		if q != "*" {
			q = toFTSQuery(q)
			query = query.Joins("JOIN (select image_id, min(rank) rank from search_index where content match ? group by image_id order by rank) t1 on t1.image_id = images.id", q)
			query = query.Order("t1.rank asc, created_at " + sortOrder)
		} else {
			query = query.Order("created_at " + sortOrder)
		}
	} else {
		query = query.Order("created_at " + sortOrder)
	}

	if q == "" {
		query = query.Where("path = ?", pathParam)
	} else if inPath && pathParam != "" {
		query = query.Where("(path = ? OR path LIKE ?)", pathParam, pathParam+"/%")
	}

	var total int64
	query.Count(&total)

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	start := int64((page - 1) * limit)

	query = query.Offset(int(start)).Limit(limit)

	var images []models.Image
	query.Find(&images)

	if directories == nil {
		directories = make([]Directory, 0)
	}

	browseImages := make([]BrowseResultImage, len(images))
	for i := range images {
		videoPreview := ""
		if strings.EqualFold(filepath.Ext(images[i].Name), ".mp4") {
			videoPreview = generateVideoPreview(&images[i])
		}
		browseImages[i] = BrowseResultImage{
			Image:        &images[i],
			VideoPreview: videoPreview,
		}
	}

	return &BrowseResult{
		Directories: directories,
		Images:      browseImages,
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

	inPath := c.Query("inPath") == "true"

	result, err := BrowseCore(pathParam, q, inPath, sortOrder, page, limit)
	if err != nil {
		if err.Error() == "Invalid path" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		}
		return
	}

	for idx, image := range result.Images {
		var imgPath string
		if image.Path != "" {
			imgPath = config.StaticImagesRoot + "/" + image.Path + "/" + image.Name
		} else {
			imgPath = config.StaticImagesRoot + "/" + image.Name
		}
		result.Images[idx].Path = appendImageToken(c, imgPath)
		if result.Images[idx].VideoPreview != "" {
			result.Images[idx].VideoPreview = appendImageToken(c, result.Images[idx].VideoPreview)
		}
	}

	c.JSON(http.StatusOK, result)
}

func UploadCore(files []*multipart.FileHeader, prefix string) ([]*models.Image, error) {
	prefix = strings.Trim(prefix, "/")

	if strings.Contains(prefix, ".") {
		return nil, errors.New("invalid prefix path")
	}

	dirName := filepath.Dir(prefix)
	if dirName == "." {
		dirName = ""
	}
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
