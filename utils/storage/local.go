package storage

import (
	"errors"
	"goventy/config"
	"goventy/utils"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/gabriel-vasile/mimetype"
	"github.com/segmentio/ksuid"
)

type LocalStorage struct {
	View      string   //public or storage
	Folder    string   //can be uploads/data or books/images
	FileName  string   //for deletion
	FileNames []string //for deletion
}

func (ls *LocalStorage) Create(file multipart.File) *UploadedFile {
	basePath := ""
	if ls.View == "storage" {
		basePath = filepath.Join(os.Getenv("goventy_ROOT"), config.ENV()["APP_ASSETS"])
	} else {
		basePath = filepath.Join(os.Getenv("goventy_ROOT"), config.ENV()["PUBLIC_ASSETS"])
	}

	if ls.Folder != "" {
		basePath = filepath.Join(basePath, ls.Folder)
	}
	//get file extension
	mime, err := mimetype.DetectReader(file)
	if err != nil {
		utils.Log(err, "error")
	}
	//generate filename
	id := ksuid.New()
	fileName := id.String() + mime.Extension()

	//create directory and parent directories if they do not exist
	os.MkdirAll(filepath.Join(basePath, ""), 0776)
	f, err := os.OpenFile(filepath.Join(basePath, fileName), os.O_CREATE|os.O_WRONLY, 0776)
	defer f.Close()
	if err != nil {
		utils.Log(err, "error")
	}
	_, err = file.Seek(0, io.SeekStart)
	io.Copy(f, file)

	uf := &UploadedFile{
		Url:        filepath.Join(ls.Folder, fileName),
		Provider:   "local",
		ProviderID: id.String(),
	}
	return uf
}

func (ls *LocalStorage) Delete(id ...string) (bool, error) {
	basePath := ""
	if ls.View == "storage" {
		basePath = filepath.Join(os.Getenv("goventy_ROOT"), config.ENV()["APP_ASSETS"])
	} else {
		basePath = filepath.Join(os.Getenv("goventy_ROOT"), config.ENV()["PUBLIC_ASSETS"])
	}

	if ls.FileName == "" {
		return false, errors.New("No file path specified")
	}

	file := filepath.Join(basePath, ls.FileName)

	err := os.Remove(file)

	if err != nil {
		utils.Log(err, "error")
		return false, err
	}

	return true, nil
}
