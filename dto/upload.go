package dto

import (
	"time"

	"github.com/google/uuid"
)

type UploadDTO struct {
	UUID uuid.UUID `json:"uuid"`
	Name string    `json:"name"`
	EXT  string    `json:"ext"`
	Size int64     `json:"size"`
	URL  string    `json:"url"`
}

func NewUploadDTO(uid uuid.UUID, name, ext string, size int64, url string) UploadDTO {
	return UploadDTO{
		EXT:  ext,
		Name: name,
		Size: size,
		UUID: uid,
		URL:  url,
	}
}

type FileDTO struct {
	UUID       uuid.UUID `json:"uuid"`
	ObjectName string    `json:"object_name"`
	Name       string    `json:"name"`
	Width      int       `json:"width"`
	Height     int       `json:"height"`
	Mime       string    `json:"mime"`
	Ext        string    `json:"ext"`
	Size       int64     `json:"size"`
	URL        string    `json:"url"`
}

type FileDTOs struct {
	UUID uuid.UUID `json:"uuid"`
	Name string    `json:"name"`
	Ext  string    `json:"ext"`
	Size int64     `json:"size"`
	URL  string    `json:"url"`

	CreatedAt time.Time `json:"created_at"`
	CreatedBy UserDTO   `json:"created_by"`
}

type ImageDTO struct {
	UUID       uuid.UUID `json:"uuid"`
	ObjectName string    `json:"object_name"`
	Name       string    `json:"name"`
	Width      int       `json:"width"`
	Height     int       `json:"height"`
	Mime       string    `json:"mime"`
	Ext        string    `json:"ext"`
	Size       int64     `json:"size"`
	URL        string    `json:"url"`
}

type ProfilePhotoDTO struct {
	Small  string `json:"small"`
	Medium string `json:"medium"`
	Large  string `json:"large"`
}
