package models

import "time"

type Video struct {
	ID              string    `json:"id"`
	UserID          string    `json:"user_id"`
	OriginalURL     string    `json:"original_url"`
	HLSURL          string    `json:"hls_url"`
	ThumbnailURL    string    `json:"thumbnail_url"`
	Duration        float64   `json:"duration"`
	Description     string    `json:"description"`
	CreatedAt       time.Time `json:"created_at"`
	Qualities       []string  `json:"qualities"`
	HLSProcessed    bool      `json:"hls_processed"`
	ProcessingError string    `json:"processing_error"`
}
