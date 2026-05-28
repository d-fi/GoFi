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

func TestParseSpotifyPlaylistEmbed(t *testing.T) {
	body := []byte(`<html><script id="__NEXT_DATA__" type="application/json">{"props":{"pageProps":{"state":{"data":{"entity":{"id":"playlist-id","title":"My Playlist","subtitle":"Owner","coverArt":{"sources":[{"url":"https://example.com/cover.jpg"}]},"trackList":[{"uri":"spotify:track:track-id","title":"Track Title","subtitle":"Artist One, Artist Two","duration":123000,"isExplicit":true}]}}}}}}</script></html>`)

	entity, err := parseSpotifyPlaylistEmbed(body)
	require.NoError(t, err)
	assert.Equal(t, "playlist-id", entity.ID)
	assert.Equal(t, "My Playlist", entity.Title)
	require.Len(t, entity.CoverArt.Sources, 1)
	assert.Equal(t, "https://example.com/cover.jpg", entity.CoverArt.Sources[0].URL)
	require.Len(t, entity.TrackList, 1)

	track := entity.TrackList[0]
	assert.Equal(t, "spotify:track:track-id", track.URI)
	assert.Equal(t, "Track Title", track.Title)
	assert.Equal(t, 123000, track.Duration)
	assert.True(t, track.IsExplicit)

	artists := spotifyArtistsFromSubtitle(track.Subtitle)
	require.Len(t, artists, 2)
	assert.Equal(t, "Artist One", artists[0].Name)
	assert.Equal(t, "Artist Two", artists[1].Name)
}
