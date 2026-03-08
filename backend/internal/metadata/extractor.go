package metadata

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/abema/go-mp4"
)

type MetadataItem struct {
	Key   string
	Value string
}

func ExtractPromptJSON(promptJSON *string) ([]MetadataItem, error) {
	var data map[string]any
	if err := json.Unmarshal([]byte(*promptJSON), &data); err != nil {
		return nil, err
	}

	var items []MetadataItem
	for _, nodeDataRaw := range data {
		nodeData, ok := nodeDataRaw.(map[string]any)
		if !ok {
			continue
		}

		inputs, ok := nodeData["inputs"].(map[string]any)
		if !ok {
			continue
		}

		for k, v := range inputs {
			if k == "type" || k == "device" {
				continue
			}

			var valStr string
			switch val := v.(type) {
			case string:
				valStr = val
			case float64:
				if val == float64(int64(val)) {
					valStr = fmt.Sprintf("%d", int64(val))
				} else {
					valStr = fmt.Sprintf("%g", val)
				}
			case bool:
				valStr = fmt.Sprintf("%v", val)
			default:
				continue
			}

			items = append(items, MetadataItem{Key: k, Value: valStr})
		}
	}

	return items, nil
}

func ExtractMetadataPNG(file *os.File) ([]MetadataItem, error) {
	var promptJSON string

	for {
		var length uint32
		if err := binary.Read(file, binary.BigEndian, &length); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		var typeBytes [4]byte
		if _, err := io.ReadFull(file, typeBytes[:]); err != nil {
			return nil, err
		}
		chunkType := string(typeBytes[:])

		if chunkType == "tEXt" {
			data := make([]byte, length)
			if _, err := io.ReadFull(file, data); err != nil {
				return nil, err
			}
			parts := bytes.SplitN(data, []byte{0}, 2)
			if len(parts) == 2 {
				key := string(parts[0])
				// ComfyUI uses 'prompt' or 'workflow'
				if key == "prompt" {
					promptJSON = string(parts[1])
				}
			}
		} else {
			// Skip data
			if _, err := file.Seek(int64(length), io.SeekCurrent); err != nil {
				return nil, err
			}
		}

		// Skip CRC
		if _, err := file.Seek(4, io.SeekCurrent); err != nil {
			return nil, err
		}

		if promptJSON != "" {
			break
		}
	}

	if promptJSON == "" {
		return nil, nil
	}

	return ExtractPromptJSON(&promptJSON)
}

func ExtractMetadataMP4(file *os.File) ([]MetadataItem, error) {
	keyIdx := -1
	ilstExpanded := -1
	promptJSON := ""

	_, err := mp4.ReadBoxStructure(file, func(h *mp4.ReadHandle) (any, error) {
		if h.BoxInfo.IsSupportedType() {
			boxType := h.BoxInfo.Type.String()
			switch boxType {
			case "keys":
				box, _, err := h.ReadPayload()
				if err != nil {
					return nil, err
				}
				keys := box.(*mp4.Keys)
				for idx, key := range keys.Entries {
					if string(key.KeyValue) == "prompt" {
						keyIdx = idx
						break
					}
				}
			case "ilst":
				ilstExpanded = 0
			default:
				if ilstExpanded >= 0 {
					if ilstExpanded == keyIdx {
						box, _, err := h.ReadPayload()
						if err != nil {
							return nil, err
						}

						item := box.(*mp4.Item)
						promptJSON = string(item.Data.Data)
						ilstExpanded = -1
					} else {
						ilstExpanded += 1
					}
				}
			}

			return h.Expand()
		}

		return nil, nil
	})

	if err != nil {
		return nil, err
	}

	if promptJSON == "" {
		return nil, nil
	}

	return ExtractPromptJSON(&promptJSON)
}

func ExtractMetadata(filepath string) ([]MetadataItem, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var header [8]byte
	if _, err := io.ReadFull(file, header[:]); err != nil {
		return nil, err
	}

	if string(header[:]) == "\x89PNG\r\n\x1a\n" {
		return ExtractMetadataPNG(file)
	} else if string(header[4:]) == "ftyp" {
		return ExtractMetadataMP4(file)
	}

	return nil, nil
}
