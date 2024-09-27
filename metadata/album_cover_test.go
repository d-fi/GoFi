package metadata

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const ALB_PICTURE = "2e018122cb56986277102d2041a592c8" // Discovery by Daft Punk

func TestDownloadAlbumCover(t *testing.T) {
	// Test valid cover sizes
	coverSizes := []int{56, 250, 500, 1000, 1500, 1800}
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
