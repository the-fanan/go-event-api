package storage

import (
	"mime/multipart"
)

type StorageInterface interface {
	Create(file multipart.File) *UploadedFile
	CreateMultiple(files []multipart.File) []*UploadedFile
	Delete(id ...string) (bool, error)
	DeleteMultiple(ids ...[]string) (bool, error)
}

type UploadedFile struct {
	Url string
	Provider string
	ProviderID string
}
