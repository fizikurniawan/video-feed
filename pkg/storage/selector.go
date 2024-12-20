package storage

import (
	"errors"
)

type MinioCfg struct {
	MINIO_ENDPOINT    string
	MINIO_ACCESS_KEY  string
	MINIO_SECRET_KEY  string
	MINIO_BUCKET_NAME string
	MINIO_IS_HTTPS    bool
}

type RackspaceCfg struct {
	RACKSPACE_USERNAME  string
	RACKSPACE_API_KEY   string
	RACKSPACE_AUTH_URL  string
	RACKSPACE_REGION    string
	RACKSPACE_CONTAINER string
}

func NewStorageService(provider string, minioCfg MinioCfg, rackCfg RackspaceCfg) (StorageService, error) {
	switch provider {
	case "minio":
		return NewMinIOService(
			minioCfg.MINIO_ENDPOINT, minioCfg.MINIO_ACCESS_KEY, minioCfg.MINIO_SECRET_KEY, minioCfg.MINIO_BUCKET_NAME, minioCfg.MINIO_IS_HTTPS,
		)
	case "rackspace":
		return NewRackspaceService(rackCfg.RACKSPACE_USERNAME, rackCfg.RACKSPACE_API_KEY, rackCfg.RACKSPACE_AUTH_URL, rackCfg.RACKSPACE_REGION, rackCfg.RACKSPACE_CONTAINER)
	default:
		return nil, errors.New("unsupported storage provider")
	}
}
