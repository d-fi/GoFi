package converter

import (
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/d-fi/GoFi/request"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	arl := os.Getenv("DEEZER_ARL")
	if arl != "" {
		if _, err := request.InitDeezerAPI(arl); err != nil {
			panic("Failed to initialize Deezer API: " + err.Error())
		}
	}
	os.Exit(m.Run())
}

func TestGetURLParts(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected URLParts
	}{
		{
			name:     "deezer track",
			url:      "https://www.deezer.com/en/track/3135556",
			expected: URLParts{ID: "3135556", Type: "track"},
		},
		{
			name:     "deezer album",
			url:      "https://www.deezer.com/en/album/6575789",
			expected: URLParts{ID: "6575789", Type: "album"},
		},
		{
			name:     "youtube watch",
			url:      "https://www.youtube.com/watch?v=qFLhGq0060w&feature=share",
			expected: URLParts{ID: "qFLhGq0060w", Type: "youtube-track"},
		},
		{
			name:     "youtube short",
			url:      "https://youtu.be/qFLhGq0060w",
			expected: URLParts{ID: "qFLhGq0060w", Type: "youtube-track"},
		},
		{
			name:     "tidal track",
			url:      "https://tidal.com/browse/track/56681096",
			expected: URLParts{ID: "56681096", Type: "tidal-track"},
		},
		{
			name:     "tidal playlist",
			url:      "https://tidal.com/browse/playlist/ed004d2b-b494-42be-8506-b1d23cd3bb80",
			expected: URLParts{ID: "ed004d2b-b494-42be-8506-b1d23cd3bb80", Type: "tidal-playlist"},
		},
		{
			name:     "spotify track",
			url:      "https://open.spotify.com/track/7FIWs0pqAYbP91WWM0vlTQ?si=abc",
			expected: URLParts{ID: "7FIWs0pqAYbP91WWM0vlTQ", Type: "spotify-track"},
		},
		{
			name:     "spotify intl album",
			url:      "https://open.spotify.com/intl-fr/album/6t7956yu5zYf5A829XRiHC",
			expected: URLParts{ID: "6t7956yu5zYf5A829XRiHC", Type: "spotify-album"},
		},
		{
			name:     "spotify uri playlist",
			url:      "spotify:playlist:37i9dQZF1DXcBWIGoYBM5M",
			expected: URLParts{ID: "37i9dQZF1DXcBWIGoYBM5M", Type: "spotify-playlist"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := GetURLParts(test.url)
			require.NoError(t, err)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetURLPartsDeezerShareLink(t *testing.T) {
	previousTransport := http.DefaultTransport
	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		require.Equal(t, http.MethodHead, req.Method)
		require.Equal(t, "link.deezer.com", req.URL.Host)

		location := "https://link.deezer.com/?dest=https%3A%2F%2Fwww.deezer.com%2Ftrack%2F3135556%3Futm_source%3Duser_sharing"
		return &http.Response{
			StatusCode: http.StatusMovedPermanently,
			Header:     http.Header{"Location": []string{location}},
			Body:       io.NopCloser(strings.NewReader("")),
			Request:    req,
		}, nil
	})
	t.Cleanup(func() {
		http.DefaultTransport = previousTransport
	})

	actual, err := GetURLParts("https://link.deezer.com/s/33mHLHCANAsGjpIDT5fji")
	require.NoError(t, err)
	assert.Equal(t, URLParts{ID: "3135556", Type: "track"}, actual)
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func TestISRCToDeezer(t *testing.T) {
	if os.Getenv("DEEZER_ARL") == "" {
		t.Skip("DEEZER_ARL is required for Deezer integration tests")
	}

	track, err := ISRCToDeezer("Harder, Better, Faster, Stronger", "GBDUW0000059")
	require.NoError(t, err)
	assert.Equal(t, "3135556", track.SNG_ID)
	assert.Equal(t, "GBDUW0000059", track.ISRC)
}

func TestUPCToDeezer(t *testing.T) {
	if os.Getenv("DEEZER_ARL") == "" {
		t.Skip("DEEZER_ARL is required for Deezer integration tests")
	}

	album, tracks, err := UPCToDeezer("Discovery", "724384960650")
	require.NoError(t, err)
	assert.Equal(t, "302127", album.ALB_ID)
	assert.Len(t, tracks, 14)
}

func TestParseInfoDeezerTrack(t *testing.T) {
	if os.Getenv("DEEZER_ARL") == "" {
		t.Skip("DEEZER_ARL is required for Deezer integration tests")
	}

	result, err := ParseInfo("https://www.deezer.com/en/track/3135556")
	require.NoError(t, err)
	assert.Equal(t, "track", result.LinkType)
	require.Len(t, result.Tracks, 1)
	assert.Equal(t, "3135556", result.Tracks[0].SNG_ID)
}

func TestYouTubeTrackToDeezer(t *testing.T) {
	if os.Getenv("DEEZER_ARL") == "" {
		t.Skip("DEEZER_ARL is required for Deezer integration tests")
	}

	track, err := YouTubeTrackToDeezer("qFLhGq0060w")
	require.NoError(t, err)
	assert.Equal(t, "136889434", track.SNG_ID)
	assert.Equal(t, "I Feel It Coming", track.SNG_TITLE)
	assert.Equal(t, "USUG11601012", track.ISRC)
}

func TestTidalTrackToDeezer(t *testing.T) {
	if os.Getenv("DEEZER_ARL") == "" {
		t.Skip("DEEZER_ARL is required for Deezer integration tests")
	}

	track, err := TidalTrackToDeezer("56681096")
	require.NoError(t, err)
	assert.Equal(t, "118190298", track.SNG_ID)
	assert.Equal(t, "QM5FT1600116", track.ISRC)
}

func TestTidalAlbumToDeezer(t *testing.T) {
	if os.Getenv("DEEZER_ARL") == "" {
		t.Skip("DEEZER_ARL is required for Deezer integration tests")
	}

	album, tracks, err := TidalAlbumToDeezer("56681092")
	require.NoError(t, err)
	assert.Equal(t, "12279688", album.ALB_ID)
	assert.Len(t, tracks, 16)
}

func TestSpotifyTrackToDeezer(t *testing.T) {
	if os.Getenv("DEEZER_ARL") == "" {
		t.Skip("DEEZER_ARL is required for Deezer integration tests")
	}

	track, err := SpotifyTrackToDeezer("7FIWs0pqAYbP91WWM0vlTQ")
	if err != nil && strings.Contains(err.Error(), "spotify API rate limited") {
		t.Skip(err.Error())
	}
	require.NoError(t, err)
	assert.Equal(t, "854914322", track.SNG_ID)
	assert.Equal(t, "USUM72000788", track.ISRC)
}
