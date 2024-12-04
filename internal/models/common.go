package models

import "time"

type ChunkInfo struct {
	UploadID    string    `json:"uploadId"`
	ChunkNumber int       `json:"chunkNumber"`
	TotalChunks int       `json:"totalChunks"`
	FileName    string    `json:"fileName"`
	CreatedAt   time.Time `json:"createdAt"`
}
