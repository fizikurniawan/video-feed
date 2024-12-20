package storage

import (
	"errors"
)

type S3Cfg struct {
	S3_ENDPOINT    string
	S3_ACCESS_KEY  string
	S3_SECRET_KEY  string
	S3_BUCKET_NAME string
	S3_IS_HTTPS    bool
}

type SwiftCfg struct {
	SWIFT_USERNAME  string
	SWIFT_API_KEY   string
	SWIFT_AUTH_URL  string
	SWIFT_REGION    string
	SWIFT_CONTAINER string
}

func NewStorageService(provider string, s3Cfg S3Cfg, swiftCfg SwiftCfg) (StorageService, error) {
	switch provider {
	case "s3":
		return NewS3Service(
			s3Cfg.S3_ENDPOINT, s3Cfg.S3_ACCESS_KEY, s3Cfg.S3_SECRET_KEY, s3Cfg.S3_BUCKET_NAME, s3Cfg.S3_IS_HTTPS,
		)
	case "swift":
		return NewSwiftService(swiftCfg.SWIFT_USERNAME, swiftCfg.SWIFT_API_KEY, swiftCfg.SWIFT_AUTH_URL, swiftCfg.SWIFT_REGION, swiftCfg.SWIFT_CONTAINER)
	default:
		return nil, errors.New("unsupported storage provider")
	}
}
