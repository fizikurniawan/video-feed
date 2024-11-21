package models

import (
	"database/sql"
	"encoding/json"
	"time"
)

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

// SaveVideo saves a video to the database or updates it if it already exists
func SaveVideo(db *sql.DB, video *Video) error {
	query := `
		INSERT INTO videos (
			id, user_id, original_url, hls_url, 
			thumbnail_url, duration, description, 
			created_at, qualities, hls_processed, 
			processing_error
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (id) DO UPDATE SET
		original_url = EXCLUDED.original_url,
		hls_url = EXCLUDED.hls_url,
		thumbnail_url = EXCLUDED.thumbnail_url,
		duration = EXCLUDED.duration,
		description = EXCLUDED.description,
		qualities = EXCLUDED.qualities,
		hls_processed = EXCLUDED.hls_processed,
		processing_error = EXCLUDED.processing_error
	`

	qualitiesJSON, err := json.Marshal(video.Qualities)
	if err != nil {
		return err
	}

	_, err = db.Exec(query,
		video.ID, video.UserID, video.OriginalURL, video.HLSURL,
		video.ThumbnailURL, video.Duration, video.Description,
		video.CreatedAt, qualitiesJSON, // SIMPAN JSON KE KOLOM JSONB
		video.HLSProcessed, video.ProcessingError,
	)
	return err
}

// GetVideoByID retrieves a video from the database by its ID
func GetVideoByID(db *sql.DB, videoID string) (*Video, error) {
	var video Video

	query := `
		SELECT 
			id, user_id, original_url, hls_url, 
			thumbnail_url, duration, description, 
			created_at, qualities, hls_processed, 
			processing_error
		FROM videos 
		WHERE id = $1
	`

	var qualitiesJSON []byte
	err := db.QueryRow(query, videoID).Scan(
		&video.ID, &video.UserID, &video.OriginalURL, &video.HLSURL,
		&video.ThumbnailURL, &video.Duration, &video.Description,
		&video.CreatedAt, &qualitiesJSON, &video.HLSProcessed,
		&video.ProcessingError,
	)

	if err != nil {
		return nil, err
	}

	// Convert JSONB ke array string
	err = json.Unmarshal(qualitiesJSON, &video.Qualities)
	if err != nil {
		return nil, err
	}

	return &video, nil
}

// UpdateVideoProcessingStatus updates the processing status of a video
func UpdateVideoProcessingStatus(db *sql.DB, videoID string, processed bool, processingError string) error {
	query := `
		UPDATE videos 
		SET hls_processed = $1, 
		    processing_error = $2 
		WHERE id = $3
	`

	_, err := db.Exec(query, processed, processingError, videoID)
	return err
}

// ListUserVideos retrieves a paginated list of videos for a specific user
func ListUserVideos(db *sql.DB, userID string, limit, offset int) ([]Video, error) {
	query := `
		SELECT 
			id, user_id, original_url, hls_url, 
			thumbnail_url, duration, description, 
			created_at, qualities, hls_processed, 
			processing_error
		FROM videos 
		WHERE user_id = $1 
		ORDER BY created_at DESC 
		LIMIT $2 OFFSET $3
	`

	rows, err := db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var videos []Video
	for rows.Next() {
		var video Video

		var qualitiesJSON []byte // JSONB akan di-scan sebagai []byte
		err = rows.Scan(
			&video.ID, &video.UserID, &video.OriginalURL, &video.HLSURL,
			&video.ThumbnailURL, &video.Duration, &video.Description,
			&video.CreatedAt, &qualitiesJSON, &video.HLSProcessed,
			&video.ProcessingError,
		)
		if err != nil {
			return nil, err
		}

		// Convert JSONB ke array string
		err = json.Unmarshal(qualitiesJSON, &video.Qualities)
		if err != nil {
			return nil, err
		}

		videos = append(videos, video)
	}

	return videos, nil
}

// DeleteVideo deletes a video by its ID
func DeleteVideo(db *sql.DB, videoID string) error {
	query := `DELETE FROM videos WHERE id = $1`
	_, err := db.Exec(query, videoID)
	return err
}
