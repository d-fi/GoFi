package converter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSpotifyTokenResource(t *testing.T) {
	tests := []struct {
		path         string
		resourceType string
		resourceID   string
	}{
		{
			path:         "tracks/7FIWs0pqAYbP91WWM0vlTQ",
			resourceType: "track",
			resourceID:   "7FIWs0pqAYbP91WWM0vlTQ",
		},
		{
			path:         "playlists/1hMzceeWw7QiI6vaBkcEJO/tracks?limit=100",
			resourceType: "playlist",
			resourceID:   "1hMzceeWw7QiI6vaBkcEJO",
		},
		{
			path:         spotifyAPIBaseURL + "albums/6t7956yu5zYf5A829XRiHC",
			resourceType: "album",
			resourceID:   "6t7956yu5zYf5A829XRiHC",
		},
	}

	for _, test := range tests {
		resourceType, resourceID, err := spotifyTokenResource(test.path)
		require.NoError(t, err)
		assert.Equal(t, test.resourceType, resourceType)
		assert.Equal(t, test.resourceID, resourceID)
	}
}

func TestSpotifyTokenResourceRejectsUnsupportedPath(t *testing.T) {
	_, _, err := spotifyTokenResource("me")
	require.Error(t, err)
}
