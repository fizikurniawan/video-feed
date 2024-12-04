package repositories

import (
	"encoding/json"
	"video-feed/internal/models"
	"video-feed/pkg/database"
)

type VideoRepository struct {
	dbManager *database.DatabaseManager
}

func NewVideoRepository(dbManager *database.DatabaseManager) *VideoRepository {
	return &VideoRepository{
		dbManager: dbManager,
	}
}

func (r *VideoRepository) Create(video *models.Video) error {
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

	// Gunakan dbManager untuk eksekusi query
	_, err = r.dbManager.Exec(query,
		video.ID, video.UserID, video.OriginalURL, video.HLSURL,
		video.ThumbnailURL, video.Duration, video.Description,
		video.CreatedAt, qualitiesJSON, // SIMPAN JSON KE KOLOM JSONB
		video.HLSProcessed, video.ProcessingError,
	)
	return err
}

func (r *VideoRepository) GetVideoByID(videoID string) (*models.Video, error) {
	var video models.Video

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
	// Gunakan dbManager untuk query
	err := r.dbManager.QueryRow(query, videoID).Scan(
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
func (r *VideoRepository) UpdateVideoProcessingStatus(videoID string, processed bool, processingError string, qualities []string, hls_url string) error {
	query := `
		UPDATE videos 
		SET hls_processed = $1, 
		    processing_error = $2,
			qualities = $3,
			hls_url = $4
		WHERE id = $5
	`

	qualitiesJSON, err := json.Marshal(qualities)
	if err != nil {
		return err
	}

	// Gunakan dbManager untuk eksekusi query
	_, err = r.dbManager.Exec(query, processed, processingError, qualitiesJSON, hls_url, videoID)
	return err
}

// ListUserVideos retrieves a paginated list of videos for a specific user
func (r *VideoRepository) ListUserVideos(userID string, limit, offset int) ([]models.Video, error) {
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

	// Menggunakan method Query dari DatabaseManager
	rows, err := r.dbManager.Query(query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var videos []models.Video
	for rows.Next() {
		var video models.Video

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
func (r *VideoRepository) DeleteVideo(videoID string) error {
	db, err := r.dbManager.GetConnection()
	if err != nil {
		return err
	}
	query := `DELETE FROM videos WHERE id = $1`
	_, err = db.Exec(query, videoID)
	return err
}
