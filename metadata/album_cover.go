package metadata

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/d-fi/GoFi/logger"
	"github.com/d-fi/GoFi/request"
	"github.com/d-fi/GoFi/utils"
	"github.com/hashicorp/golang-lru/v2/expirable"
)

const (
	cacheSize = 50
	cacheTTL  = 30 * time.Minute
)

const DefaultCoverFileName = "cover.jpg"

func IsValidCoverSize(size int) bool {
	switch size {
	case 56, 250, 500, 1000, 1200, 1400, 1500, 1800:
		return true
	default:
		return false
	}
}

func NormalizeCoverSize(size, fallback int) int {
	if IsValidCoverSize(size) {
		return size
	}
	return fallback
}

var albumCoverCache = expirable.NewLRU[string, []byte](cacheSize, nil, cacheTTL)

// DownloadAlbumCover downloads an album cover based on the provided album picture hash and cover size.
func DownloadAlbumCover(albumPicture string, albumCoverSize int) ([]byte, error) {
	logger.Debug("Attempting to download album cover with hash: %s and size: %d", albumPicture, albumCoverSize)

	if albumPicture == "" {
		logger.Debug("Album picture hash is empty.")
		return nil, errors.New("album picture hash is empty")
	}

	if !IsValidCoverSize(albumCoverSize) {
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

func NormalizeCoverFileName(fileName string) string {
	fileName = strings.TrimSpace(filepath.Base(fileName))
	if fileName == "." || fileName == string(filepath.Separator) {
		fileName = ""
	}
	if fileName == "" {
		return DefaultCoverFileName
	}
	ext := strings.ToLower(filepath.Ext(fileName))
	base := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	if ext == "" {
		fileName += ".jpg"
	} else if ext != ".jpg" && ext != ".jpeg" {
		fileName = base + ".jpg"
	}
	fileName = utils.SanitizeFileName(fileName)
	if fileName == "" || fileName == "." {
		return DefaultCoverFileName
	}
	return fileName
}

func SaveAlbumCoverFile(dir string, fileName string, albumPicture string, albumCoverSize int) (string, error) {
	cover, err := DownloadAlbumCover(albumPicture, albumCoverSize)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	path := filepath.Join(dir, NormalizeCoverFileName(fileName))
	if existing, err := os.ReadFile(path); err == nil {
		if bytes.Equal(existing, cover) {
			return path, nil
		}
		logger.Debug("Skipping cover file because %s already exists with different data", path)
		return "", nil
	}
	if err := os.WriteFile(path, cover, 0644); err != nil {
		return "", err
	}
	return path, nil
}
