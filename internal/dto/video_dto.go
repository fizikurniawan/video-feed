package dto

import "mime/multipart"

type InitiateChunkDTO struct {
	FileName    string `json:"fileName"`
	TotalChunks int    `json:"totalChunks"`
}

type ChunkUploadDTO struct {
	UploadID    string                `form:"uploadId" binding:"required"`
	ChunkNumber string                `form:"chunkNumber" binding:"required"`
	Chunk       *multipart.FileHeader `form:"chunk" binding:"required"`
}

type CompleteChunkUploadDTO struct {
	UploadID    string `json:"uploadId" binding:"required"`
	Description string `json:"description"`
}
