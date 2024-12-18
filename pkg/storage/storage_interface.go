package storage

import "os"

type StorageService interface {
	UploadObject(objectName string, file *os.File) error
}
