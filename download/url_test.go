package download

import (
	"os"
	"strconv"
	"testing"

	"github.com/d-fi/GoFi/api"
	"github.com/d-fi/GoFi/request"
	"github.com/stretchr/testify/assert"
)

const (
	SNG_ID = "3135556" // Harder, Better, Faster, Stronger by Daft Punk
)

func init() {
	// Initialize the Deezer API for all tests
	arl := os.Getenv("DEEZER_ARL")
	_, err := request.InitDeezerAPI(arl)
	if err != nil {
		panic("Failed to initialize Deezer API: " + err.Error())
	}
}

func TestDzAuthenticate(t *testing.T) {
	user, err := DzAuthenticate()
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.NotEmpty(t, user.LicenseToken)
	assert.True(t, user.CanStreamLossless || user.CanStreamHQ)
	assert.NotEmpty(t, user.Country)
}

func TestGetTrackUrlFromServer(t *testing.T) {
	trackToken := "example_track_token"
	_, err := GetTrackUrlFromServer(trackToken, "MP3_320")
	assert.Error(t, err, "Expected error due to incorrect token or unavailable track")
}

func TestGetTrackDownloadUrl(t *testing.T) {
	track, err := api.GetTrackInfo(SNG_ID)
	assert.NoError(t, err, "Failed to fetch track information")
	assert.NotEmpty(t, track.MD5_ORIGIN, "MD5 origin should not be empty")
	assert.NotEmpty(t, track.TRACK_TOKEN, "Track token should not be empty")

	// Testing various qualities
	qualities := []int{1, 3, 9}

	for _, quality := range qualities {
		t.Run("Quality "+strconv.Itoa(quality), func(t *testing.T) {
			trackURL, err := GetTrackDownloadUrl(track, quality)
			if err == nil {
				assert.NotNil(t, trackURL)
				assert.NotEmpty(t, trackURL.TrackUrl)
				assert.Greater(t, trackURL.FileSize, int(0))
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "Your account can't stream")
			}
		})
	}
}

func TestGetTrackDownloadUrlWithInvalidQuality(t *testing.T) {
	track, err := api.GetTrackInfo(SNG_ID)
	assert.NoError(t, err, "Failed to fetch track information")
	assert.NotEmpty(t, track.TRACK_TOKEN, "Track token should not be empty")

	_, err = GetTrackDownloadUrl(track, 999) // Testing an invalid quality
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown quality 999")
}
