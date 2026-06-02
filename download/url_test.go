package download

import (
	"context"
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

var hasDeezerARL bool

func TestMain(m *testing.M) {
	arl := os.Getenv("DEEZER_ARL")
	if arl == "" {
		os.Exit(m.Run())
	}

	_, err := request.InitDeezerAPI(arl)
	if err != nil {
		panic("Failed to initialize Deezer API: " + err.Error())
	}

	hasDeezerARL = true
	os.Exit(m.Run())
}

func requireDeezerARL(t *testing.T) {
	t.Helper()
	if !hasDeezerARL {
		t.Skip("DEEZER_ARL is required")
	}
}

func TestParseDeezerUserDataAllowsNullCapabilityFields(t *testing.T) {
	user, err := parseDeezerUserData([]byte(`{
		"results": {
			"COUNTRY": "US",
			"USER": {
				"OPTIONS": {
					"license_token": "token",
					"web_lossless": null,
					"mobile_loseless": true,
					"web_hq": null,
					"mobile_hq": false
				}
			}
		}
	}`))
	assert.NoError(t, err)
	assert.Equal(t, "token", user.LicenseToken)
	assert.True(t, user.CanStreamLossless)
	assert.False(t, user.CanStreamHQ)
	assert.Equal(t, "US", user.Country)
}

func TestParseDeezerUserDataRequiresLicenseToken(t *testing.T) {
	_, err := parseDeezerUserData([]byte(`{"results":{"USER":{"OPTIONS":{}}}}`))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing license token")
}

func TestDzAuthenticate(t *testing.T) {
	requireDeezerARL(t)
	user, err := DzAuthenticate(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.NotEmpty(t, user.LicenseToken)
	assert.True(t, user.CanStreamLossless || user.CanStreamHQ)
	assert.NotEmpty(t, user.Country)
}

func TestGetTrackUrlFromServer(t *testing.T) {
	requireDeezerARL(t)
	trackToken := "example_track_token"
	_, err := GetTrackUrlFromServer(context.Background(), trackToken, "MP3_320")
	assert.Error(t, err, "Expected error due to incorrect token or unavailable track")
}

func TestGetTrackDownloadUrl(t *testing.T) {
	requireDeezerARL(t)
	track, err := api.GetTrackInfo(SNG_ID)
	assert.NoError(t, err, "Failed to fetch track information")
	assert.NotEmpty(t, track.MD5_ORIGIN, "MD5 origin should not be empty")
	assert.NotEmpty(t, track.TRACK_TOKEN, "Track token should not be empty")

	// Testing various qualities
	qualities := []int{1, 3, 9}

	for _, quality := range qualities {
		t.Run("Quality "+strconv.Itoa(quality), func(t *testing.T) {
			trackURL, err := GetTrackDownloadUrl(context.Background(), track, quality)
			if err == nil {
				assert.NotNil(t, trackURL)
				assert.NotEmpty(t, trackURL.TrackUrl)
				assert.Greater(t, trackURL.FileSize, int64(0))
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "Your account can't stream")
			}
		})
	}
}

func TestGetTrackDownloadUrlWithInvalidQuality(t *testing.T) {
	requireDeezerARL(t)
	track, err := api.GetTrackInfo(SNG_ID)
	assert.NoError(t, err, "Failed to fetch track information")
	assert.NotEmpty(t, track.TRACK_TOKEN, "Track token should not be empty")

	_, err = GetTrackDownloadUrl(context.Background(), track, 999) // Testing an invalid quality
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown quality 999")
}
