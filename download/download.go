package download

import (
	"fmt"
	"os"
	"path/filepath"
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

	// Download the track from the generated URL
	resp, err := request.Client.R().Get(trackData.TrackUrl)
	if err != nil {
		logger.Debug("Failed to download track: %v", err)
		return "", fmt.Errorf("failed to download track: %v", err)
	}
	logger.Debug("Track downloaded successfully")

	// Decrypt the downloaded track if necessary
	trackBody := resp.Body()
	if trackData.IsEncrypted {
		logger.Debug("Track is encrypted, starting decryption process")
		trackBody = decrypt.DecryptDownload(resp.Body(), track.SNG_ID)
		logger.Debug("Track decrypted successfully")
	}

	// Add metadata to the downloaded track
	trackWithMetadata, err := metadata.AddTrackTags(trackBody, track, options.CoverSize)
	if err != nil {
		logger.Debug("Failed to add metadata: %v", err)
		return "", fmt.Errorf("failed to add metadata: %v", err)
	}
	logger.Debug("Metadata added successfully")

	// Write the track to the specified directory
	if err := os.WriteFile(savedPath, trackWithMetadata, 0644); err != nil {
		logger.Debug("Failed to save track: %v", err)
		return "", fmt.Errorf("failed to save track: %v", err)
	}
	logger.Debug("Track saved to: %s", savedPath)

	return savedPath, nil
}
