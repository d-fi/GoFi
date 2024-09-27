package metadata

import (
	"errors"
	"fmt"
	"time"

	"github.com/d-fi/GoFi/logger"
	"github.com/d-fi/GoFi/request"
	"github.com/hashicorp/golang-lru/v2/expirable"
)

const (
	cacheSize = 50
	cacheTTL  = 30 * time.Minute
)

// Valid cover sizes
const (
	CoverSize56   = 56
	CoverSize250  = 250
	CoverSize500  = 500
	CoverSize1000 = 1000
	CoverSize1500 = 1500
	CoverSize1800 = 1800
)

var validCoverSizes = map[int]bool{
	CoverSize56:   true,
	CoverSize250:  true,
	CoverSize500:  true,
	CoverSize1000: true,
	CoverSize1500: true,
	CoverSize1800: true,
}

var albumCoverCache = expirable.NewLRU[string, []byte](cacheSize, nil, cacheTTL)

// DownloadAlbumCover downloads an album cover based on the provided album picture hash and cover size.
func DownloadAlbumCover(albumPicture string, albumCoverSize int) ([]byte, error) {
	logger.Debug("Attempting to download album cover with hash: %s and size: %d", albumPicture, albumCoverSize)

	if albumPicture == "" {
		logger.Debug("Album picture hash is empty.")
		return nil, errors.New("album picture hash is empty")
	}

	if !validCoverSizes[albumCoverSize] {
		logger.Debug("Invalid cover size requested: %d", albumCoverSize)
		return nil, fmt.Errorf("invalid cover size: %d", albumCoverSize)
	}

	cacheKey := fmt.Sprintf("%s%d", albumPicture, albumCoverSize)
	if cachedData, ok := albumCoverCache.Get(cacheKey); ok {
		logger.Debug("Album cover retrieved from cache: %s", cacheKey)
		return cachedData, nil
	}

	url := fmt.Sprintf("https://e-cdns-images.dzcdn.net/images/cover/%s/%dx%d-000000-80-0-0.jpg",
		albumPicture, albumCoverSize, albumCoverSize)
	logger.Debug("Downloading album cover from URL: %s", url)

	resp, err := request.Client.R().Get(url)
	if err != nil {
		logger.Debug("Failed to download album cover: %v", err)
		return nil, fmt.Errorf("failed to download album cover: %w", err)
	}

	data := resp.Body()
	albumCoverCache.Add(cacheKey, data)
	logger.Debug("Album cover downloaded and cached successfully: %s", cacheKey)

	return data, nil
}
