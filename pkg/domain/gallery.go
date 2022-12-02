package domain

import (
	"fmt"
	"time"
)

func GetGalleryPrefix() string {
	return fmt.Sprintf("%s:gallery", publicPrefix)
}

func GetGalleryAttributeKey(attribute string) string {
	return fmt.Sprintf("%s:%s", GetGalleryPrefix(), attribute)
}

func GetSortedGalleryKey() string {
	return fmt.Sprintf("%s:sorted", GetGalleryPrefix())
}

func GetAllGalleryKey(name string) string {
	return fmt.Sprintf("%s:name:%s", GetGalleryPrefix(), name)
}

type MetadataJsonDate struct {
	ID        int64      `json:"-"      gorm:"primaryKey;auto_increment"` // 自增主键，自增主键不能有任何业务含义。
	CreatedAt *time.Time `json:"-"      gorm:"not null;index"`
	UpdatedAt *time.Time `json:"-"      gorm:"not null;index"`
	URI       string     `json:"uri"    gorm:"uniqueIndex"`
	Data      string     `json:"data"   gorm:"type:text"`
}
