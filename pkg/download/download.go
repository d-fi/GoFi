package download

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/d-fi/GoFi/pkg/api"
	"github.com/d-fi/GoFi/pkg/decrypt"
	"github.com/d-fi/GoFi/pkg/types"
	"github.com/go-resty/resty/v2"
)

var client = resty.New()

// Path to save music, update this to match your configuration
var MusicPath = "./music" // Change this to the desired path for storing music files

// DownloadTrack downloads a track, adds metadata, and saves it to disk.
func DownloadTrack(sngID string, quality int, ext string, coverSize int) (string, error) {
	// Fetch track information
	track, err := api.GetTrackInfo(sngID)
	if err != nil {
		return "", fmt.Errorf("failed to fetch track info: %v", err)
	}

	// Create directory if it does not exist
	qualityPath := filepath.Join(MusicPath, fmt.Sprintf("%d", quality))
	if _, err := os.Stat(qualityPath); os.IsNotExist(err) {
		if err := os.MkdirAll(qualityPath, 0755); err != nil {
			return "", fmt.Errorf("failed to create directory: %v", err)
		}
	}

	// Get the track download URL
	trackData, err := GetTrackDownloadUrl(track, quality)
	if err != nil || trackData == nil {
		return "", fmt.Errorf("failed to retrieve downloadable URL: %v", err)
	}

	// Set up the save path for the track
	safeTitle := strings.ReplaceAll(track.SNG_TITLE, "/", "_")
	savedPath := filepath.Join(qualityPath, fmt.Sprintf("%s-%s.%s", safeTitle, track.SNG_ID, ext))

	// Check if the file exists and update its timestamp if it does
	if _, err := os.Stat(savedPath); err == nil {
		currentTime := time.Now().Local()
		if err := os.Chtimes(savedPath, currentTime, currentTime); err != nil {
			return "", fmt.Errorf("failed to update file timestamps: %v", err)
		}
	} else {
		// Download the track
		resp, err := client.R().Get(trackData.TrackUrl)
		if err != nil {
			return "", fmt.Errorf("failed to download track: %v", err)
		}

		// Check if decryption is needed and add metadata
		var trackBody []byte
		if trackData.IsEncrypted {
			trackBody = decrypt.DecryptDownload(resp.Body(), track.SNG_ID)
		} else {
			trackBody = resp.Body()
		}

		// Add metadata to the track
		trackWithMetadata, err := addTrackTags(trackBody, track, coverSize)
		if err != nil {
			return "", fmt.Errorf("failed to add metadata: %v", err)
		}

		// Write the track to disk
		if err := os.WriteFile(savedPath, trackWithMetadata, 0644); err != nil {
			return "", fmt.Errorf("failed to save track: %v", err)
		}
	}

	return savedPath, nil
}

// addTrackTags adds metadata to the track. Implement this function based on your metadata requirements.
func addTrackTags(body []byte, track types.TrackType, coverSize int) ([]byte, error) {
	// Add track metadata handling code here
	// This is a placeholder implementation. Replace with actual metadata tagging logic.
	return body, nil
}
