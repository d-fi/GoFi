package converter

import (
	"fmt"
	"strings"

	"github.com/d-fi/GoFi/request"
	"github.com/d-fi/GoFi/types"
)

const tidalBaseURL = "https://api.tidal.com/v1/"

type TidalArtist struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type TidalAlbumRef struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Cover string `json:"cover"`
}

type TidalTrack struct {
	ID                   int           `json:"id"`
	Title                string        `json:"title"`
	Duration             int           `json:"duration"`
	PremiumStreamingOnly bool          `json:"premiumStreamingOnly"`
	TrackNumber          int           `json:"trackNumber"`
	Copyright            string        `json:"copyright"`
	URL                  string        `json:"url"`
	Explicit             bool          `json:"explicit"`
	AudioQuality         string        `json:"audioQuality"`
	Artist               TidalArtist   `json:"artist"`
	Album                TidalAlbumRef `json:"album"`
	ISRC                 string        `json:"isrc"`
	Editable             bool          `json:"editable"`
}

type TidalAlbum struct {
	ID                   int         `json:"id"`
	Title                string      `json:"title"`
	Duration             int         `json:"duration"`
	PremiumStreamingOnly bool        `json:"premiumStreamingOnly"`
	TrackNumber          int         `json:"trackNumber"`
	Copyright            string      `json:"copyright"`
	URL                  string      `json:"url"`
	Explicit             bool        `json:"explicit"`
	AudioQuality         string      `json:"audioQuality"`
	Artist               TidalArtist `json:"artist"`
	Cover                string      `json:"cover"`
	VideoCover           *string     `json:"videoCover"`
	UPC                  string      `json:"upc"`
}

type TidalPlaylist struct {
	UUID           string `json:"uuid"`
	Title          string `json:"title"`
	NumberOfTracks int    `json:"numberOfTracks"`
	NumberOfVideos int    `json:"numberOfVideos"`
	Creator        struct {
		ID int `json:"id"`
	} `json:"creator"`
	Description    string `json:"description"`
	Duration       int    `json:"duration"`
	LastUpdated    string `json:"lastUpdated"`
	Created        string `json:"created"`
	Type           string `json:"type"`
	PublicPlaylist bool   `json:"publicPlaylist"`
	URL            string `json:"url"`
	Image          string `json:"image"`
}

type tidalList[T any] struct {
	Limit              int `json:"limit"`
	Offset             int `json:"offset"`
	TotalNumberOfItems int `json:"totalNumberOfItems"`
	Items              []T `json:"items"`
}

type AlbumArtURLs struct {
	Small  string `json:"sm"`
	Medium string `json:"md"`
	Large  string `json:"lg"`
	XL     string `json:"xl"`
}

// GetTidalTrack fetches a Tidal track by id.
func GetTidalTrack(id string) (TidalTrack, error) {
	var track TidalTrack
	err := tidalGet("tracks/"+id, &track)
	return track, err
}

// TidalTrackToDeezer converts a Tidal track to a Deezer track via ISRC.
func TidalTrackToDeezer(id string) (types.TrackType, error) {
	track, err := GetTidalTrack(id)
	if err != nil {
		return types.TrackType{}, err
	}
	return ISRCToDeezer(track.Title, track.ISRC)
}

// GetTidalAlbum fetches a Tidal album by id.
func GetTidalAlbum(id string) (TidalAlbum, error) {
	var album TidalAlbum
	err := tidalGet("albums/"+id, &album)
	return album, err
}

// TidalAlbumToDeezer converts a Tidal album to a Deezer album and track list via UPC.
func TidalAlbumToDeezer(id string) (types.AlbumType, []types.TrackType, error) {
	album, err := GetTidalAlbum(id)
	if err != nil {
		return types.AlbumType{}, nil, err
	}
	return UPCToDeezer(album.Title, album.UPC)
}

// GetTidalAlbumTracks fetches tracks for a Tidal album id.
func GetTidalAlbumTracks(id string) ([]TidalAlbum, int, error) {
	var list tidalList[TidalAlbum]
	err := tidalGet("albums/"+id+"/tracks", &list)
	return list.Items, list.TotalNumberOfItems, err
}

// GetTidalArtistAlbums fetches Tidal artist albums filtered to the requested artist id.
func GetTidalArtistAlbums(id string) ([]TidalAlbum, int, error) {
	var list tidalList[TidalAlbum]
	if err := tidalGet("artists/"+id+"/albums", &list); err != nil {
		return nil, 0, err
	}

	filtered := make([]TidalAlbum, 0, len(list.Items))
	for _, item := range list.Items {
		if fmt.Sprintf("%d", item.Artist.ID) == id {
			filtered = append(filtered, item)
		}
	}
	return filtered, list.TotalNumberOfItems, nil
}

// GetTidalArtistTopTracks fetches Tidal artist top tracks filtered to the requested artist id.
func GetTidalArtistTopTracks(id string) ([]TidalTrack, int, error) {
	var list tidalList[TidalTrack]
	if err := tidalGet("artists/"+id+"/toptracks", &list); err != nil {
		return nil, 0, err
	}

	filtered := make([]TidalTrack, 0, len(list.Items))
	for _, item := range list.Items {
		if fmt.Sprintf("%d", item.Artist.ID) == id {
			filtered = append(filtered, item)
		}
	}
	return filtered, list.TotalNumberOfItems, nil
}

// GetTidalPlaylist fetches a Tidal playlist by uuid.
func GetTidalPlaylist(uuid string) (TidalPlaylist, error) {
	var playlist TidalPlaylist
	err := tidalGet("playlists/"+uuid, &playlist)
	return playlist, err
}

// GetTidalPlaylistTracks fetches Tidal playlist tracks by uuid.
func GetTidalPlaylistTracks(uuid string) ([]TidalTrack, int, error) {
	var list tidalList[TidalTrack]
	err := tidalGet("playlists/"+uuid+"/tracks", &list)
	return list.Items, list.TotalNumberOfItems, err
}

// TidalAlbumArtToURL builds album art URLs for a Tidal cover uuid.
func TidalAlbumArtToURL(uuid string) AlbumArtURLs {
	baseURL := "https://resources.tidal.com/images/" + strings.ReplaceAll(uuid, "-", "/")
	return AlbumArtURLs{
		Small:  baseURL + "/160x160.jpg",
		Medium: baseURL + "/320x320.jpg",
		Large:  baseURL + "/640x640.jpg",
		XL:     baseURL + "/1280x1280.jpg",
	}
}

// TidalArtistToDeezer converts Tidal artist top tracks to Deezer tracks.
func TidalArtistToDeezer(id string) ([]types.TrackType, error) {
	items, _, err := GetTidalArtistTopTracks(id)
	if err != nil {
		return nil, err
	}

	tracks := make([]types.TrackType, 0, len(items))
	for _, item := range items {
		track, err := ISRCToDeezer(item.Title, item.ISRC)
		if err != nil {
			continue
		}
		tracks = append(tracks, track)
	}
	return tracks, nil
}

// TidalPlaylistToDeezer converts a Tidal playlist to Deezer playlist metadata and matching Deezer tracks.
func TidalPlaylistToDeezer(uuid string) (types.PlaylistInfo, []types.TrackType, error) {
	body, err := GetTidalPlaylist(uuid)
	if err != nil {
		return types.PlaylistInfo{}, nil, err
	}
	items, _, err := GetTidalPlaylistTracks(uuid)
	if err != nil {
		return types.PlaylistInfo{}, nil, err
	}

	tracks := make([]types.TrackType, 0, len(items))
	for index, item := range items {
		track, err := ISRCToDeezer(item.Title, item.ISRC)
		if err != nil {
			continue
		}
		position := index + 1
		track.TRACK_POSITION = &position
		tracks = append(tracks, track)
	}

	userID := fmt.Sprintf("%d", body.Creator.ID)
	playlist := types.PlaylistInfo{
		PlaylistID:      body.UUID,
		Description:     body.Description,
		ParentUsername:  userID,
		ParentUserID:    userID,
		PictureType:     "cover",
		PlaylistPicture: body.Image,
		Title:           body.Title,
		Type:            "0",
		Status:          0,
		UserID:          userID,
		DateAdd:         body.Created,
		DateMod:         body.LastUpdated,
		DateCreate:      body.Created,
		NbSong:          body.NumberOfTracks,
		NbFan:           0,
		Checksum:        body.Created,
		HasArtistLinked: false,
		IsSponsored:     false,
		IsEdito:         false,
		TYPE_INTERNAL:   "playlist",
	}
	return playlist, tracks, nil
}

func tidalGet(path string, target interface{}) error {
	resp, err := request.Client.R().
		SetResult(target).
		SetHeader("user-agent", "TIDAL/3704 CFNetwork/1220.1 Darwin/20.3.0").
		SetHeader("x-tidal-token", "i4ZDjcyhed7Mu47q").
		SetQueryParam("limit", "500").
		SetQueryParam("countryCode", "US").
		Get(tidalBaseURL + path)
	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("tidal API error: %s", resp.Status())
	}
	return nil
}
