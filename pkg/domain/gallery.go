package domain

import (
	"fmt"
)

var GalleryPrefix = fmt.Sprintf("%s:gallery", publicPrefix)

func GetGalleryAttributeKey(attribute string) string {
	return fmt.Sprintf("%s:%s", GalleryPrefix, attribute)
}

func GetSortedGalleryKey() string {
	return fmt.Sprintf("%s:sorted", GalleryPrefix)
}

func GetAllGalleryKey(name string) string {
	return fmt.Sprintf("%s:name:%s", GalleryPrefix, name)
}
