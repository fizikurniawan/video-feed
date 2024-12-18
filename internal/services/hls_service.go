package services

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"video-feed/config"
	"video-feed/internal/repositories"
	"video-feed/pkg/storage"
)

type HLSBackgroundJob struct {
	Cfg     *config.AppConfig
	Storage storage.StorageService
	Repo    *repositories.VideoRepository
}

func NewHLSBackgroundJob(Cfg *config.AppConfig, Storage storage.StorageService, Repo *repositories.VideoRepository) *HLSBackgroundJob {
	return &HLSBackgroundJob{Cfg: Cfg, Storage: Storage, Repo: Repo}
}

type HLSJobResult struct {
	VideoID string
	Success bool
	Error   error
}

type Resolution struct {
	Name    string
	Height  int
	Bitrate string
}

var resolutions = []Resolution{
	{Name: "480p", Height: 480, Bitrate: "1000k"},
	{Name: "720p", Height: 720, Bitrate: "2500k"},
}

func (h *HLSBackgroundJob) ProcessHLSWithTimeout(videoID, inputPath string) <-chan HLSJobResult {
	resultChan := make(chan HLSJobResult, 1)

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()
		defer close(resultChan)

		result := HLSJobResult{VideoID: videoID}

		// Create output directory
		outputDir := fmt.Sprintf("tmp/videos/%s", videoID)
		os.MkdirAll(outputDir, os.ModePerm)

		// Defer cleanup of output directory
		defer func() {
			if err := os.RemoveAll(outputDir); err != nil {
				log.Printf("Failed to cleanup directory %s: %v", outputDir, err)
			}
		}()

		// Create master playlist
		masterPlaylist := "#EXTM3U\n#EXT-X-VERSION:3\n"

		// Process each resolution
		for _, res := range resolutions {
			resPath := filepath.Join(outputDir, res.Name)
			os.MkdirAll(resPath, os.ModePerm)

			if err := h.processQuality(ctx, inputPath, resPath, res.Height, res.Bitrate); err != nil {
				result.Error = fmt.Errorf("%s conversion failed: %v", res.Name, err)
				result.Success = false
				resultChan <- result
				return
			}

			masterPlaylist += fmt.Sprintf("#EXT-X-STREAM-INF:BANDWIDTH=%s,RESOLUTION=%dp\n%s/playlist.m3u8\n",
				strings.TrimSuffix(res.Bitrate, "k")+"000",
				res.Height,
				res.Name)
		}

		// Save master playlist
		masterPlaylistPath := filepath.Join(outputDir, "master.m3u8")
		if err := os.WriteFile(masterPlaylistPath, []byte(masterPlaylist), 0644); err != nil {
			result.Error = fmt.Errorf("failed to write master playlist: %v", err)
			result.Success = false
			resultChan <- result
			return
		}

		// Upload all files to MinIO
		err := filepath.Walk(outputDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() && (strings.HasSuffix(path, ".ts") || strings.HasSuffix(path, ".m3u8")) {
				file, err := os.Open(path)
				if err != nil {
					return err
				}
				defer file.Close()

				relativePath := strings.TrimPrefix(path, outputDir)
				minioPath := fmt.Sprintf("videos/%s%s", videoID, relativePath)
				return h.Storage.UploadObject(minioPath, file)
			}
			return nil
		})

		if err != nil {
			result.Error = fmt.Errorf("upload HLS segments failed: %v", err)
			result.Success = false
		} else {
			result.Success = true
		}

		resultChan <- result
	}()

	return resultChan
}

func (h *HLSBackgroundJob) processQuality(ctx context.Context, inputPath, outputDir string, height int, bitrate string) error {
	playlistPath := filepath.Join(outputDir, "playlist.m3u8")
	segmentPath := filepath.Join(outputDir, "segment%03d.ts")

	args := []string{
		"-i", inputPath,
		"-profile:v", "baseline",
		"-level", "3.0",
		"-start_number", "0",
		"-hls_time", "10",
		"-hls_list_size", "0",
		"-f", "hls",
	}

	if height > 0 {
		args = append(args,
			"-vf", fmt.Sprintf("scale=-2:%d", height),
			"-b:v", bitrate,
		)
	}

	args = append(args,
		"-hls_segment_filename", segmentPath,
		playlistPath,
	)

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("FFmpeg failed: %v, output: %s", err, output)
	}

	return nil
}

func (h *HLSBackgroundJob) HandleJobResult(result HLSJobResult) error {
	var (
		processingError string
		qualities       = []string{"original"} // Default qualities
		hlsURL          string
	)

	// Capture error message if any
	if result.Error != nil {
		processingError = result.Error.Error()
	}

	// Handle success case
	if result.Success {
		qualities = []string{"480p", "720p", "original"} // Update qualities
		// Generate HLS URL
		hlsURL = h.Cfg.Env.CDN_URL + "videos/" + result.VideoID + "/master.m3u8"
	}

	// Update video processing status
	return h.Repo.UpdateVideoProcessingStatus(result.VideoID, result.Success, processingError, qualities, hlsURL)
}
