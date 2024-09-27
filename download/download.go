package download

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/d-fi/GoFi/api"
	"github.com/d-fi/GoFi/decrypt"
	"github.com/d-fi/GoFi/logger"
	"github.com/d-fi/GoFi/metadata"
	"github.com/d-fi/GoFi/request"
	"github.com/d-fi/GoFi/utils"
)

// DownloadTrack downloads a track, adds metadata, and saves it to the specified directory.
func DownloadTrack(options TrackDownloadOptions) (string, error) {
	logger.Debug("Starting download for track ID: %s with quality: %d", options.SngID, options.Quality)
	track, err := api.GetTrackInfo(options.SngID)
	if err != nil {
		logger.Debug("Failed to fetch track info: %v", err)
		return "", fmt.Errorf("failed to fetch track info: %v", err)
	}

	// Create directory for saving the track if it does not exist
	if err := os.MkdirAll(options.SaveToDir, 0755); err != nil {
		logger.Debug("Failed to create directory: %v", err)
		return "", fmt.Errorf("failed to create directory: %v", err)
	}
	logger.Debug("Directory created: %s", options.SaveToDir)

	trackData, err := GetTrackDownloadUrl(track, options.Quality)
	if err != nil || trackData == nil {
		logger.Debug("Failed to retrieve downloadable URL: %v", err)
		return "", fmt.Errorf("failed to retrieve downloadable URL: %v", err)
	}
	logger.Debug("Download URL retrieved: %s", trackData.TrackUrl)

	// Sanitize the track title to ensure it's safe for file systems
	safeTitle := utils.SanitizeFileName(track.SNG_TITLE)
	savedPath := filepath.Join(options.SaveToDir, fmt.Sprintf("%s-%s.%s", safeTitle, track.SNG_ID, options.Ext))
	logger.Debug("Saving track as: %s", savedPath)

	// If the file exists, update its timestamp and return the path
	if _, err := os.Stat(savedPath); err == nil {
		if err := os.Chtimes(savedPath, time.Now(), time.Now()); err != nil {
			logger.Debug("Failed to update file timestamps: %v", err)
			return "", fmt.Errorf("failed to update file timestamps: %v", err)
		}
		logger.Debug("File already exists, updated timestamp: %s", savedPath)
		return savedPath, nil
	}

	// Open the destination file
	out, err := os.Create(savedPath)
	if err != nil {
		logger.Debug("Failed to create file: %v", err)
		return "", fmt.Errorf("failed to create file: %v", err)
	}
	defer func() {
		out.Close()
		if err != nil {
			logger.Debug("Removing incomplete file: %s", savedPath)
			_ = os.Remove(savedPath)
		}
	}()

	// Set up signal handling for interrupt (Ctrl+C) to clean up the file
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Channel to notify when download is complete
	done := make(chan struct{})

	go func() {
		select {
		case <-interrupt:
			logger.Debug("Interrupt signal received, removing incomplete file: %s", savedPath)
			_ = os.Remove(savedPath)
			os.Exit(1)
		case <-done:
			// Download completed normally
		}
	}()

	// Download the track from the generated URL with progress tracking
	resp, err := request.Client.R().
		SetDoNotParseResponse(true). // Do not parse the response to handle the stream manually
		Get(trackData.TrackUrl)

	if err != nil {
		logger.Debug("Failed to download track: %v", err)
		return "", fmt.Errorf("failed to download track: %v", err)
	}
	defer resp.RawBody().Close()

	logger.Debug("Track download started")

	// Get the total size of the file from the response headers
	contentLength := resp.RawResponse.ContentLength

	// Track download progress
	buffer := make([]byte, 32*1024) // 32 KB buffer size
	var totalBytesRead int64

	for {
		n, readErr := resp.RawBody().Read(buffer)
		if n > 0 {
			_, writeErr := out.Write(buffer[:n])
			if writeErr != nil {
				logger.Debug("Failed to write to file: %v", writeErr)
				return "", fmt.Errorf("failed to write to file: %v", writeErr)
			}

			totalBytesRead += int64(n)

			// Call the progress callback if provided
			if options.OnProgress != nil && contentLength > 0 {
				progress := float64(totalBytesRead) / float64(contentLength) * 100
				options.OnProgress(progress, totalBytesRead, contentLength)
			}
		}

		if readErr == io.EOF {
			break
		}

		if readErr != nil {
			logger.Debug("Failed during download: %v", readErr)
			_ = os.Remove(savedPath)
			return "", fmt.Errorf("failed during download: %v", readErr)
		}
	}

	// Notify that the download is complete
	close(done)

	logger.Debug("Track downloaded successfully")

	// Decrypt the downloaded track if necessary
	trackBody, err := os.ReadFile(savedPath)
	if err != nil {
		logger.Debug("Failed to read downloaded file for decryption: %v", err)
		return "", fmt.Errorf("failed to read downloaded file for decryption: %v", err)
	}

	if trackData.IsEncrypted {
		logger.Debug("Track is encrypted, starting decryption process")
		trackBody = decrypt.DecryptDownload(trackBody, track.SNG_ID)
		logger.Debug("Track decrypted successfully")
	}

	// Add metadata to the downloaded track
	trackWithMetadata, err := metadata.AddTrackTags(trackBody, track, options.CoverSize)
	if err != nil {
		logger.Debug("Failed to add metadata: %v", err)
		return "", fmt.Errorf("failed to add metadata: %v", err)
	}
	logger.Debug("Metadata added successfully")

	// Write the track with metadata back to the specified file
	if err := os.WriteFile(savedPath, trackWithMetadata, 0644); err != nil {
		logger.Debug("Failed to save track with metadata: %v", err)
		return "", fmt.Errorf("failed to save track with metadata: %v", err)
	}
	logger.Debug("Track saved with metadata to: %s", savedPath)

	return savedPath, nil
}
