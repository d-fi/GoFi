package api

import (
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/d-fi/GoFi/request"
	"github.com/stretchr/testify/assert"
)

const (
	SNG_ID = "3135556" // Harder, Better, Faster, Stronger by Daft Punk
	ALB_ID = "302127"  // Discovery by Daft Punk
)

func init() {
	// Initialize the Deezer API for all tests
	arl := os.Getenv("DEEZER_ARL")
	_, err := request.InitDeezerAPI(arl)
	if err != nil {
		panic("Failed to initialize Deezer API: " + err.Error())
	}
}

func TestGetUser(t *testing.T) {
	response, err := GetUser()
	assert.NoError(t, err)
	assert.NotEmpty(t, response.BlogName)
	assert.NotEmpty(t, response.Email)
	assert.NotEmpty(t, response.UserID)
	assert.Equal(t, "user", response.Type)
}

func TestGetTrackInfo(t *testing.T) {
	response, err := GetTrackInfo(SNG_ID)
	assert.NoError(t, err)
	assert.Equal(t, SNG_ID, response.SNG_ID)
	assert.Equal(t, "GBDUW0000059", response.ISRC)
	assert.Equal(t, "000790eceb6cb6732d225c0585632b31", response.MD5_ORIGIN)
	assert.Equal(t, "song", response.Type)
}

func TestGetTrackInfoPublicApi(t *testing.T) {
	response, err := GetTrackInfoPublicApi(SNG_ID)
	assert.NoError(t, err)
	assert.Equal(t, SNG_ID, strconv.Itoa(response.ID))
	assert.Equal(t, "GBDUW0000059", response.ISRC)
	assert.Equal(t, "track", response.Type)
}

func TestGetLyrics(t *testing.T) {
	response, err := GetLyrics(SNG_ID)
	assert.NoError(t, err)
	assert.NotNil(t, response.LYRICS_ID)
	assert.Equal(t, "2780622", *response.LYRICS_ID)
	assert.Greater(t, len(response.LYRICS_TEXT), 0)
}

func TestGetAlbumInfo(t *testing.T) {
	response, err := GetAlbumInfo(ALB_ID)
	assert.NoError(t, err)
	assert.Equal(t, ALB_ID, response.ALB_ID)
	assert.Equal(t, "724384960650", response.UPC)
	assert.Equal(t, "album", response.TYPE_INTERNAL)
}

func TestGetAlbumInfoPublicApi(t *testing.T) {
	response, err := GetAlbumInfoPublicApi(ALB_ID)
	assert.NoError(t, err)
	assert.Equal(t, ALB_ID, strconv.Itoa(response.ID))
	assert.Equal(t, "724384960650", response.UPC)
	assert.Equal(t, "album", response.Type)
}

func TestGetAlbumTracks(t *testing.T) {
	response, err := GetAlbumTracks(ALB_ID)
	assert.NoError(t, err)
	assert.Equal(t, 14, response.Count)
	assert.Equal(t, len(response.Data), response.Count)
}

func TestGetPlaylistInfo(t *testing.T) {
	PLAYLIST_ID := "4523119944"
	response, err := GetPlaylistInfo(PLAYLIST_ID)
	assert.NoError(t, err)
	assert.Greater(t, response.NbSong, 0)
	assert.Equal(t, "sayem314", response.ParentUsername)
	assert.Equal(t, "playlist", response.TYPE_INTERNAL)
}

func TestGetPlaylistTracks(t *testing.T) {
	PLAYLIST_ID := "4523119944"
	response, err := GetPlaylistTracks(PLAYLIST_ID)
	assert.NoError(t, err)
	assert.Greater(t, response.Count, 0)
	assert.Equal(t, len(response.Data), response.Count)
}

func TestGetArtistInfo(t *testing.T) {
	ART_ID := "13"
	response, err := GetArtistInfo(ART_ID)
	assert.NoError(t, err)
	assert.Equal(t, "Eminem", response.ART_NAME)
	assert.Equal(t, "artist", response.TYPE_INTERNAL)
}

func TestGetDiscography(t *testing.T) {
	ART_ID := "13"
	response, err := GetDiscography(ART_ID, 10)
	assert.NoError(t, err)
	assert.Equal(t, 10, response.Count)
	assert.Equal(t, len(response.Data), response.Count)
}

func TestGetProfile(t *testing.T) {
	USER_ID := "2064440442"
	response, err := GetProfile(USER_ID)
	assert.NoError(t, err)
	assert.Equal(t, "sayem314", response.USER.BLOG_NAME)
	assert.Equal(t, "user", response.USER.TYPE_INTERNAL)
}

func TestSearchAlternative(t *testing.T) {
	ARTIST := "Eminem"
	TRACK := "The Real Slim Shady"
	response, err := SearchAlternative(ARTIST, TRACK, 10)
	assert.NoError(t, err)
	assert.Equal(t, "artist:'eminem' track:'the real slim shady'", response.QUERY)
	assert.Equal(t, len(response.TRACK.Data), response.TRACK.Count)
}

func TestSearchMusic(t *testing.T) {
	QUERY := "Eminem"
	response, err := SearchMusic(QUERY, 1, "TRACK", "ALBUM", "ARTIST")
	assert.NoError(t, err)
	assert.Equal(t, strings.ToLower(QUERY), response.QUERY)
	assert.Greater(t, response.TRACK.Count, 0)
	assert.Greater(t, response.ALBUM.Count, 0)
	assert.Greater(t, response.ARTIST.Count, 0)
}

func TestGetChannelList(t *testing.T) {
	response, err := GetChannelList()
	assert.NoError(t, err)
	assert.Greater(t, response.Count, 0)
	assert.Equal(t, len(response.Data), response.Count)
}

func TestGetShowInfo(t *testing.T) {
	response, err := GetShowInfo("338532", 10, 0)
	assert.NoError(t, err)
	assert.Equal(t, "201952", response.Data.LabelID)
	assert.Equal(t, 10, response.Episodes.Count)
	assert.True(t, len(response.Episodes.Data) > 0)
}
