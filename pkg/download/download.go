package download

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/d-fi/GoFi/pkg/api"
	"github.com/d-fi/GoFi/pkg/decrypt"
	"github.com/d-fi/GoFi/pkg/request"
	"github.com/d-fi/GoFi/pkg/types"
	"github.com/d-fi/GoFi/pkg/utils"
)

// DownloadTrack downloads a track, adds metadata, and saves it to the specified directory.
func DownloadTrack(options TrackDownloadOptions) (string, error) {
	track, err := api.GetTrackInfo(options.SngID)
	if err != nil {
		return "", fmt.Errorf("failed to fetch track info: %v", err)
	}

	// Create directory for saving the track if it does not exist
	if err := os.MkdirAll(options.SaveToDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %v", err)
	}

	trackData, err := GetTrackDownloadUrl(track, options.Quality)
	if err != nil || trackData == nil {
		return "", fmt.Errorf("failed to retrieve downloadable URL: %v", err)
	}

	// Sanitize the track title to ensure it's safe for file systems
	safeTitle := utils.SanitizeFileName(track.SNG_TITLE)
	savedPath := filepath.Join(options.SaveToDir, fmt.Sprintf("%s-%s.%s", safeTitle, track.SNG_ID, options.Ext))

	// If the file exists, update its timestamp and return the path
	if _, err := os.Stat(savedPath); err == nil {
		if err := os.Chtimes(savedPath, time.Now(), time.Now()); err != nil {
			return "", fmt.Errorf("failed to update file timestamps: %v", err)
		}
		return savedPath, nil
	}

	// Download the track from the generated URL
	resp, err := request.Client.R().Get(trackData.TrackUrl)
	if err != nil {
		return "", fmt.Errorf("failed to download track: %v", err)
	}

	// Decrypt the downloaded track if necessary
	trackBody := resp.Body()
	if trackData.IsEncrypted {
		trackBody = decrypt.DecryptDownload(resp.Body(), track.SNG_ID)
	}

	// Add metadata to the downloaded track
	trackWithMetadata, err := addTrackTags(trackBody, track, options.CoverSize)
	if err != nil {
		return "", fmt.Errorf("failed to add metadata: %v", err)
	}

	// Write the track to the specified directory
	if err := os.WriteFile(savedPath, trackWithMetadata, 0644); err != nil {
		return "", fmt.Errorf("failed to save track: %v", err)
	}

	return savedPath, nil
}

// addTrackTags adds metadata to the track. Placeholder for actual tagging logic.
func addTrackTags(body []byte, track types.TrackType, coverSize int) ([]byte, error) {
	// Implement metadata tagging logic here
	return body, nil
}
