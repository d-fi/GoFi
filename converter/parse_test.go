package converter

import (
	"os"
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
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := GetURLParts(test.url)
			require.NoError(t, err)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetURLPartsUnsupported(t *testing.T) {
	_, err := GetURLParts("https://open.spotify.com/track/3UmaczJpikHgJFyBTAJVoz")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "spotify URLs are not supported")
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
