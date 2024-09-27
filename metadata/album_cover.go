package metadata

import (
	"errors"
	"fmt"
	"time"

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

func DownloadAlbumCover(albumPicture string, albumCoverSize int) ([]byte, error) {
	if albumPicture == "" {
		return nil, errors.New("album picture hash is empty")
	}

	if !validCoverSizes[albumCoverSize] {
		return nil, fmt.Errorf("invalid cover size: %d", albumCoverSize)
	}

	cacheKey := fmt.Sprintf("%s%d", albumPicture, albumCoverSize)
	if cachedData, ok := albumCoverCache.Get(cacheKey); ok {
		return cachedData, nil
	}

	url := fmt.Sprintf("https://e-cdns-images.dzcdn.net/images/cover/%s/%dx%d-000000-80-0-0.jpg",
		albumPicture, albumCoverSize, albumCoverSize)

	resp, err := request.Client.R().Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download album cover: %w", err)
	}
	data := resp.Body()

	albumCoverCache.Add(cacheKey, data)
	return data, nil
}
