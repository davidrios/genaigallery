package handlers

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gorm.io/gorm"

	"genai-gallery-backend/internal/config"
	"genai-gallery-backend/internal/database"
	"genai-gallery-backend/internal/models"
	gallerysync "genai-gallery-backend/internal/sync"
)

func setupTestDB(t *testing.T) (*gorm.DB, func()) {
	t.Helper()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Could not determine cwd: %v", err)
	}

	originalImagesDir := config.ImagesDir
	originalDBPath := config.DBPath

	config.ImagesDir = filepath.Join(cwd, "..", "..", "testdata", "fixtures")
	config.DBPath = filepath.Join(cwd, "..", "..", "testdata", "fixtures", "handlers_test.db")

	os.Remove(config.DBPath)

	database.InitDB(config.DBPath)
	db := database.GetDB()

	gallerysync.PerformSync(db)

	teardown := func() {
		config.ImagesDir = originalImagesDir
		config.DBPath = originalDBPath
	}

	return db, teardown
}

func TestGetImageCore(t *testing.T) {
	db, teardown := setupTestDB(t)
	defer teardown()

	var img models.Image
	err := db.First(&img).Error
	if err != nil {
		t.Fatalf("Failed to retrieve a base image from DB: %v", err)
	}

	fetchedImg, err := GetImageCore(int(img.ID))
	if err != nil {
		t.Fatalf("GetImageCore failed for valid ID %d: %v", img.ID, err)
	}

	if fetchedImg.ID != img.ID {
		t.Errorf("Expected image ID %d, got %d", img.ID, fetchedImg.ID)
	}

	if fetchedImg.Path != img.Path {
		t.Errorf("Expected image Path %s, got %s", img.Path, fetchedImg.Path)
	}

	// Test Invalid ID
	_, err = GetImageCore(9999999)
	if err == nil {
		t.Error("Expected an error for non-existent image ID 9999999, but got nil")
	}
}

func TestBrowseCore(t *testing.T) {
	db, teardown := setupTestDB(t)
	defer teardown()

	var initialCount int64
	db.Model(&models.Image{}).Count(&initialCount)

	if initialCount != 5 {
		t.Fatalf("Database have incorrect image count: %v", initialCount)
	}

	t.Run("Browse Root Directory", func(t *testing.T) {
		res, err := BrowseCore("", "", false, "asc", 1, 50)
		if err != nil {
			t.Fatalf("BrowseCore failed: %v", err)
		}

		if res.Total != 4 {
			t.Errorf("Expected 4 images in root, got %v", res.Total)
		}

		// Root should see the 'video' directory based on our fixtures map
		hasVideoDir := false
		for _, dir := range res.Directories {
			if dir.Name == "video" {
				hasVideoDir = true
				break
			}
		}

		if !hasVideoDir {
			t.Errorf("Expected to see 'video' directory in browse results")
		}
	})

	t.Run("Browse Root Directory paginated", func(t *testing.T) {
		res, err := BrowseCore("", "", false, "asc", 1, 3)
		if err != nil {
			t.Fatalf("BrowseCore failed: %v", err)
		}

		if res.Total != 4 {
			t.Errorf("Expected 4 images in root, got %v", res.Total)
		}

		if res.Pages != 2 {
			t.Errorf("Expected 2 pages in root, got %v", res.Total)
		}

		if len(res.Images) != 3 {
			t.Errorf("Expected 3 images in page 1, got %v", res.Total)
		}

		// Root should see the 'video' directory based on our fixtures map
		hasVideoDir := false
		for _, dir := range res.Directories {
			if dir.Name == "video" {
				hasVideoDir = true
				break
			}
		}

		if !hasVideoDir {
			t.Errorf("Expected to see 'video' directory in browse results")
		}

		res, err = BrowseCore("", "", false, "asc", 2, 3)
		if err != nil {
			t.Fatalf("BrowseCore failed: %v", err)
		}

		if len(res.Images) != 1 {
			t.Errorf("Expected 1 images in page 2, got %v", res.Total)
		}
	})

	t.Run("Browse Subdirectory", func(t *testing.T) {
		res, err := BrowseCore("video", "", false, "asc", 1, 50)
		if err != nil {
			t.Fatalf("BrowseCore failed for 'video' subdir: %v", err)
		}

		if len(res.Images) != 0 {
			t.Errorf("Expected 0 media files within 'video' directory, got %d", len(res.Images))
		}

		if len(res.Directories) != 1 {
			t.Fatalf("Expected 1 subdirectory within 'video', got %d", len(res.Directories))
		}

		dir := res.Directories[0]
		if dir.Name != "subfolder" {
			t.Errorf("Expected subdirectory 'subfolder', got %v", dir.Name)
		}

		if strings.Contains(dir.Path, "\\") {
			t.Errorf("Directory path must use forward slashes, got: %s", dir.Path)
		}

		if dir.Path != "video/subfolder" {
			t.Errorf("Expected directory path to be 'video/subfolder', got: %s", dir.Path)
		}
	})

	t.Run("Browse Sub-subdirectory", func(t *testing.T) {
		res, err := BrowseCore("video/subfolder", "", false, "asc", 1, 50)
		if err != nil {
			t.Fatalf("BrowseCore failed for 'video/subfolder' subdir: %v", err)
		}

		if len(res.Images) != 1 {
			t.Fatalf("Expected 1 media file within 'video/subfolder' directory")
		}

		img := res.Images[0]

		if strings.Contains(img.Path, "\\") {
			t.Errorf("Image path must use forward slashes, got: %s", img.Path)
		}

		if img.Path != "video/subfolder" {
			t.Errorf("Image path %s not matched to 'video/subfolder' dir", img.Path)
		}
	})

	t.Run("Browse FTS Query", func(t *testing.T) {
		tests := []struct {
			query    string
			expected int
		}{
			{"hidre", 1},
			{"wan2.2", 1},
			{"qwen", 3},
			{"lora_name:qwen", 2},
			{"qwen fp8", 2},
			{"hidream OR wan2.2", 2},
			{"qwen alien", 2},
		}

		for _, tt := range tests {
			t.Run(tt.query, func(t *testing.T) {
				res, err := BrowseCore("", tt.query, false, "asc", 1, 50)
				if err != nil {
					t.Fatalf("BrowseCore FTS query failed: %v", err)
				}

				if res.Total != int64(tt.expected) {
					t.Fatalf("Expected %v results, got %v", tt.expected, res.Total)
				}
			})
		}

		t.Run("Search while in path with inPath false", func(t *testing.T) {
			res, err := BrowseCore("video", "qwen", false, "asc", 1, 50)
			if err != nil {
				t.Fatalf("BrowseCore FTS query failed: %v", err)
			}

			if res.Total != 3 {
				t.Fatalf("Expected 3 results, got %v", res.Total)
			}
		})

		t.Run("In path 1", func(t *testing.T) {
			res, err := BrowseCore("", "wan2.2", true, "asc", 1, 50)
			if err != nil {
				t.Fatalf("BrowseCore FTS query failed: %v", err)
			}

			if res.Total != 0 {
				t.Fatalf("Expected 0 results, got %v", res.Total)
			}
		})

		t.Run("In path 2", func(t *testing.T) {
			res, err := BrowseCore("video/subfolder", "qwen OR wan2.2", true, "asc", 1, 50)
			if err != nil {
				t.Fatalf("BrowseCore FTS query failed: %v", err)
			}

			if res.Total != 1 {
				t.Fatalf("Expected 1 result, got %v", res.Total)
			}

			if res.Images[0].Name != "ComfyUI_00001_.mp4" {
				t.Fatalf("Expected %v result, got %v", "ComfyUI_00001_.mp4", res.Images[0].Name)
			}
		})
	})
}

func TestUploadCore(t *testing.T) {
	db, teardown := setupTestDB(t)
	defer teardown()

	fakeImageContent := []byte("fake image content")

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("files", "test_upload.png")
	if err != nil {
		t.Fatalf("Could not create form file: %v", err)
	}
	_, err = part.Write(fakeImageContent)
	if err != nil {
		t.Fatalf("Could not write image content: %v", err)
	}
	writer.Close()

	req, err := http.NewRequest("POST", "/upload", body)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	err = req.ParseMultipartForm(10 << 20)
	if err != nil {
		t.Fatalf("Failed to parse multipart form: %v", err)
	}

	files := req.MultipartForm.File["files"]
	if len(files) == 0 {
		t.Fatalf("No files parsed from multipart form")
	}

	t.Run("Test upload to root", func(t *testing.T) {
		var initialCount int64
		db.Model(&models.Image{}).Count(&initialCount)

		images, err := UploadCore(files, "ComfyUI")
		if err != nil {
			t.Fatalf("UploadCore failed: %v", err)
		}

		defer func() {
			for _, image := range images {
				imagePath := filepath.Join(config.ImagesDir, image.Name)
				os.Remove(imagePath)
			}
		}()

		if len(images) != 1 {
			t.Fatal("Expected 1 image")
		}

		var newCount int64
		db.Model(&models.Image{}).Count(&newCount)
		if newCount <= initialCount {
			t.Errorf("Expected image count to increase after upload, went from %d to %d", initialCount, newCount)
		}

		imagePath := filepath.Join(config.ImagesDir, images[0].Name)

		content, err := os.ReadFile(imagePath)
		if err != nil {
			t.Fatalf("Could not read uploaded image: %v", err)
		}

		if !bytes.Equal(content, fakeImageContent) {
			t.Fatalf("Unexpected image content. Expected '%v' found '%v'", fakeImageContent, content)
		}
	})

	t.Run("Test upload to subdir", func(t *testing.T) {
		var initialCount int64
		db.Model(&models.Image{}).Count(&initialCount)

		images, err := UploadCore(files, "a/sub/dir/ComfyUI")
		if err != nil {
			t.Fatalf("UploadCore failed: %v", err)
		}

		defer func() {
			for _, image := range images {
				imagePath := filepath.Join(config.ImagesDir, filepath.FromSlash(image.Path), image.Name)
				os.Remove(imagePath)
			}
		}()

		if len(images) != 1 {
			t.Fatal("Expected 1 image")
		}

		var newCount int64
		db.Model(&models.Image{}).Count(&newCount)
		if newCount <= initialCount {
			t.Errorf("Expected image count to increase after upload, went from %d to %d", initialCount, newCount)
		}

		imagePath := filepath.Join(config.ImagesDir, filepath.FromSlash(images[0].Path), images[0].Name)
		expectedPath := filepath.Join(config.ImagesDir, "a", "sub", "dir", images[0].Name)
		if expectedPath != imagePath {
			t.Fatalf("Saved path %v different than expected %v", imagePath, expectedPath)
		}

		content, err := os.ReadFile(imagePath)
		if err != nil {
			t.Fatalf("Could not read uploaded image: %v", err)
		}

		if !bytes.Equal(content, fakeImageContent) {
			t.Fatalf("Unexpected image content. Expected '%v' found '%v'", fakeImageContent, content)
		}
	})
}
