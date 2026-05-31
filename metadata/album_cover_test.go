package metadata

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/d-fi/GoFi/request"
	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const ALB_PICTURE = "2e018122cb56986277102d2041a592c8" // Discovery by Daft Punk

func TestDownloadAlbumCover(t *testing.T) {
	// Test valid cover sizes
	coverSizes := []int{56, 250, 500, 1000, 1200, 1400, 1500, 1800}
	for _, size := range coverSizes {
		cover, err := DownloadAlbumCover(ALB_PICTURE, size)
		assert.NoError(t, err)
		assert.NotNil(t, cover)
		assert.Greater(t, len(cover), 0)
	}

	// Test invalid cover sizes
	_, err := DownloadAlbumCover(ALB_PICTURE, 2000)
	assert.Error(t, err)
	assert.Equal(t, "invalid cover size: 2000", err.Error())

	// Test empty album picture hash
	_, err = DownloadAlbumCover("", 500)
	assert.Error(t, err)
	assert.Equal(t, "album picture hash is empty", err.Error())
}

func TestDownloadAlbumCoverDoesNotCacheHTTPError(t *testing.T) {
	previousClient := request.Client
	previousCache := albumCoverCache
	previousAlbumCoverURL := albumCoverURL
	t.Cleanup(func() {
		request.Client = previousClient
		albumCoverCache = previousCache
		albumCoverURL = previousAlbumCoverURL
	})

	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if requests == 1 {
			http.Error(w, "blocked", http.StatusForbidden)
			return
		}
		_, err := w.Write([]byte("jpeg"))
		require.NoError(t, err)
	}))
	t.Cleanup(server.Close)

	request.Client = resty.New()
	albumCoverCache = expirable.NewLRU[string, []byte](cacheSize, nil, cacheTTL)
	albumCoverURL = func(albumPicture string, albumCoverSize int) string {
		return server.URL
	}

	_, err := DownloadAlbumCover(ALB_PICTURE, 500)
	require.Error(t, err)
	require.Contains(t, err.Error(), "403 Forbidden")

	cover, err := DownloadAlbumCover(ALB_PICTURE, 500)
	require.NoError(t, err)
	assert.Equal(t, []byte("jpeg"), cover)
	assert.Equal(t, 2, requests)
}

func TestIsValidCoverSize(t *testing.T) {
	tests := map[int]bool{
		49:   false,
		50:   true,
		56:   true,
		1200: true,
		1234: true,
		1400: true,
		1800: true,
		1801: false,
	}
	for size, expected := range tests {
		if got := IsValidCoverSize(size); got != expected {
			t.Fatalf("IsValidCoverSize(%d) = %v, want %v", size, got, expected)
		}
	}
}
