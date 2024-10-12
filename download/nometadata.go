package download

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/d-fi/GoFi/api"
	"github.com/d-fi/GoFi/decrypt"
	"github.com/d-fi/GoFi/logger"
	"github.com/d-fi/GoFi/request"
)

// DownloadTrackWithoutMetadata downloads a track, decrypts if necessary, and returns the buffer without adding metadata.
func DownloadTrackWithoutMetadata(ctx context.Context, options DownloadTrackWithoutMetadataOptions) ([]byte, error) {
	logger.Debug("Starting download for track ID: %s with quality: %d (no metadata)", options.SngID, options.Quality)

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

	req := request.Client.R().
		SetDoNotParseResponse(true).
		SetContext(ctx)

	resp, err := req.Get(trackData.TrackUrl)
	if err != nil {
		logger.Debug("Failed to download track: %v", err)
		return nil, fmt.Errorf("failed to download track: %v", err)
	}
	defer resp.RawBody().Close()

	logger.Debug("Track download started")

	var buffer bytes.Buffer
	_, err = io.Copy(&buffer, resp.RawBody())
	if err != nil {
		logger.Debug("Failed during download: %v", err)
		return nil, fmt.Errorf("failed during download: %v", err)
	}

	logger.Debug("Track downloaded successfully")

	trackBody := buffer.Bytes()
	if trackData.IsEncrypted {
		logger.Debug("Track is encrypted, starting decryption process")
		trackBody = decrypt.DecryptDownload(trackBody, track.SNG_ID)
		logger.Debug("Track decrypted successfully")
	}

	return trackBody, nil
}
