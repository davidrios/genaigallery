package models

import (
	"time"
)

type Image struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Path      string    `gorm:"type:varchar(4096);uniqueIndex:path_name_uniq" json:"path"`
	Name      string    `gorm:"type:varchar(1024);uniqueIndex:path_name_uniq" json:"name"`
	CreatedAt time.Time `json:"created_at"`

	MetadataItems []ImageMetadata `gorm:"foreignKey:ImageID;constraint:OnDelete:CASCADE" json:"metadata_items"`
}

type ImageMetadata struct {
	ID      uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	ImageID uint   `gorm:"index" json:"image_id"`
	Key     string `gorm:"type:varchar(255);index" json:"key"`
	Value   string `gorm:"type:text;index" json:"value"`
}
