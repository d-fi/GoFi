package download

import (
	"bytes"
	"fmt"
	"io"

	"github.com/d-fi/GoFi/api"
	"github.com/d-fi/GoFi/decrypt"
	"github.com/d-fi/GoFi/logger"
	"github.com/d-fi/GoFi/metadata"
	"github.com/d-fi/GoFi/request"
)

// DownloadTrackToBuffer downloads a track, decrypts if necessary, adds metadata, and returns the buffer.
func DownloadTrackToBuffer(options DownloadTrackToBufferOptions) ([]byte, error) {
	logger.Debug("Starting download for track ID: %s with quality: %d", options.SngID, options.Quality)
	track, err := api.GetTrackInfo(options.SngID)
	if err != nil {
		logger.Debug("Failed to fetch track info: %v", err)
		return nil, fmt.Errorf("failed to fetch track info: %v", err)
	}

	trackData, err := GetTrackDownloadUrl(track, options.Quality)
	if err != nil || trackData == nil {
		logger.Debug("Failed to retrieve downloadable URL: %v", err)
		return nil, fmt.Errorf("failed to retrieve downloadable URL: %v", err)
	}
	logger.Debug("Download URL retrieved: %s", trackData.TrackUrl)

	// Download the track from the generated URL without saving to disk
	resp, err := request.Client.R().
		SetDoNotParseResponse(true).
		Get(trackData.TrackUrl)

	if err != nil {
		logger.Debug("Failed to download track: %v", err)
		return nil, fmt.Errorf("failed to download track: %v", err)
	}
	defer resp.RawBody().Close()

	logger.Debug("Track download started")

	// Buffer to store the downloaded content
	var buffer bytes.Buffer
	_, err = io.Copy(&buffer, resp.RawBody())
	if err != nil {
		logger.Debug("Failed during download: %v", err)
		return nil, fmt.Errorf("failed during download: %v", err)
	}

	logger.Debug("Track downloaded successfully")

	// Read the buffer to decrypt if necessary
	trackBody := buffer.Bytes()
	if trackData.IsEncrypted {
		logger.Debug("Track is encrypted, starting decryption process")
		trackBody = decrypt.DecryptDownload(trackBody, track.SNG_ID)
		logger.Debug("Track decrypted successfully")
	}

	// Add metadata to the downloaded track
	trackWithMetadata, err := metadata.AddTrackTags(trackBody, track, options.CoverSize)
	if err != nil {
		logger.Debug("Failed to add metadata: %v", err)
		return nil, fmt.Errorf("failed to add metadata: %v", err)
	}
	logger.Debug("Metadata added successfully")

	return trackWithMetadata, nil
}
