package metadata

import (
	"encoding/binary"
	"os"
	"testing"
)

func createDummyPNG(t *testing.T, chunks map[string]string) string {
	t.Helper()

	f, err := os.CreateTemp("", "test_*.png")
	if err != nil {
		t.Fatal(err)
	}

	// PNG signature
	_, err = f.Write([]byte("\x89PNG\r\n\x1a\n"))
	if err != nil {
		t.Fatal(err)
	}

	for chunkType, data := range chunks {
		length := uint32(len(data))
		err = binary.Write(f, binary.BigEndian, length)
		if err != nil {
			t.Fatal(err)
		}
		_, err = f.Write([]byte(chunkType))
		if err != nil {
			t.Fatal(err)
		}
		_, err = f.Write([]byte(data))
		if err != nil {
			t.Fatal(err)
		}
		// Fake CRC
		err = binary.Write(f, binary.BigEndian, uint32(0))
		if err != nil {
			t.Fatal(err)
		}
	}

	// Write IEND
	err = binary.Write(f, binary.BigEndian, uint32(0))
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.Write([]byte("IEND"))
	if err != nil {
		t.Fatal(err)
	}
	err = binary.Write(f, binary.BigEndian, uint32(0))
	if err != nil {
		t.Fatal(err)
	}

	name := f.Name()
	f.Close()
	return name
}

func TestExtractMetadata(t *testing.T) {
	t.Run("ValidComfyUIPNG", func(t *testing.T) {
		pngPath := "../../testdata/fixtures/ComfyUI_00001_.png"

		items, err := ExtractMetadata(pngPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(items) == 0 {
			t.Fatalf("expected items extracted, got 0")
		}

		// Spot check a few known metadata keys that should exist in a comfyUI image
		expectedKeys := []string{"seed", "steps", "cfg", "sampler_name"}
		foundMap := make(map[string]bool)

		for _, item := range items {
			foundMap[item.Key] = true
		}

		for _, expectedKey := range expectedKeys {
			if !foundMap[expectedKey] {
				t.Errorf("expected to find metadata key %s, but didn't", expectedKey)
			}
		}
	})

	t.Run("ValidComfyUIMP4", func(t *testing.T) {
		pngPath := "../../testdata/fixtures/video/ComfyUI_00001_.mp4"

		items, err := ExtractMetadata(pngPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(items) == 0 {
			t.Fatalf("expected items extracted, got 0")
		}

		// Spot check a few known metadata keys that should exist in a comfyUI image
		expectedKeys := []string{"seed", "steps", "cfg", "sampler_name"}
		foundMap := make(map[string]bool)

		for _, item := range items {
			foundMap[item.Key] = true
		}

		for _, expectedKey := range expectedKeys {
			if !foundMap[expectedKey] {
				t.Errorf("expected to find metadata key %s, but didn't", expectedKey)
			}
		}
	})

	t.Run("FileNotFound", func(t *testing.T) {
		_, err := ExtractMetadata("non_existent_file.png")
		if err == nil {
			t.Fatal("expected error for non-existent file, got nil")
		}
	})

	t.Run("NotAPNG", func(t *testing.T) {
		f, err := os.CreateTemp("", "test_*.jpg")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(f.Name())

		_, err = f.Write([]byte("not a png file content"))
		if err != nil {
			t.Fatal(err)
		}
		f.Close()

		items, err := ExtractMetadata(f.Name())
		if err != nil {
			t.Fatalf("expected no error for non-png, got: %v", err)
		}
		if items != nil {
			t.Fatalf("expected nil items for non-png, got: %v", items)
		}
	})

	t.Run("MissingPromptChunk", func(t *testing.T) {
		pngPath := createDummyPNG(t, map[string]string{
			"tEXt": "other\x00data",
		})
		defer os.Remove(pngPath)

		items, err := ExtractMetadata(pngPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if items != nil {
			t.Fatalf("expected nil items, got: %v", items)
		}
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		pngPath := createDummyPNG(t, map[string]string{
			"tEXt": "prompt\x00{invalid json",
		})
		defer os.Remove(pngPath)

		_, err := ExtractMetadata(pngPath)
		if err == nil {
			t.Fatal("expected json unmarshal error, got nil")
		}
	})

	t.Run("MalformedPNG", func(t *testing.T) {
		f, err := os.CreateTemp("", "test_*.png")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(f.Name())

		// Write valid signature but EOF immediately
		_, err = f.Write([]byte("\x89PNG\r\n\x1a\n"))
		if err != nil {
			t.Fatal(err)
		}
		f.Close()

		items, err := ExtractMetadata(f.Name())
		if err != nil {
			if err.Error() != "EOF" {
				t.Logf("expected EOF, got: %v", err)
			}
		} else if len(items) != 0 {
			t.Fatalf("expected 0 items for malformed png, got: %v", items)
		}
	})
}
