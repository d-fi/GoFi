package converter

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/d-fi/GoFi/types"
)

const (
	spotifyAPIBaseURL       = "https://api.spotify.com/v1/"
	spotifyBrowserUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0 Safari/537.36"
	spotifyPartnerQueryURL  = "https://api-partner.spotify.com/pathfinder/v1/query"
	spotifyPlaylistQuery    = "queryPlaylist"
	spotifyPlaylistQuerySHA = "908a5597b4d0af0489a9ad6a2d41bc3b416ff47c0884016d92bbd6822d0eb6d8"
	spotifyPartnerPageLimit = 1000
)

var (
	spotifyTokenMu     sync.Mutex
	spotifyToken       string
	spotifyTokenKey    string
	spotifyTokenExpiry time.Time
	spotifyHTTPClient  = &http.Client{Timeout: 15 * time.Second}
)

type SpotifyArtist struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type SpotifyImage struct {
	URL    string `json:"url"`
	Height int    `json:"height"`
	Width  int    `json:"width"`
}

type SpotifyAlbumRef struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	AlbumType   string            `json:"album_type"`
	ReleaseDate string            `json:"release_date"`
	Images      []SpotifyImage    `json:"images"`
	Artists     []SpotifyArtist   `json:"artists"`
	ExternalIDs map[string]string `json:"external_ids"`
	TotalTracks int               `json:"total_tracks"`
	Type        string            `json:"type"`
	URI         string            `json:"uri"`
}

type SpotifyTrack struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	DurationMS  int               `json:"duration_ms"`
	Explicit    bool              `json:"explicit"`
	Artists     []SpotifyArtist   `json:"artists"`
	Album       SpotifyAlbumRef   `json:"album"`
	ExternalIDs map[string]string `json:"external_ids"`
	Type        string            `json:"type"`
	URI         string            `json:"uri"`
}

type SpotifyAlbum struct {
	SpotifyAlbumRef
	Copyrights []struct {
		Text string `json:"text"`
		Type string `json:"type"`
	} `json:"copyrights"`
}

type SpotifyOwner struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	Type        string `json:"type"`
}

type SpotifyPlaylist struct {
	ID            string         `json:"id"`
	Name          string         `json:"name"`
	Description   string         `json:"description"`
	Public        bool           `json:"public"`
	Collaborative bool           `json:"collaborative"`
	Images        []SpotifyImage `json:"images"`
	Owner         SpotifyOwner   `json:"owner"`
	Tracks        struct {
		Total int `json:"total"`
	} `json:"tracks"`
	Type string `json:"type"`
	URI  string `json:"uri"`
}

type spotifyList[T any] struct {
	Items []T    `json:"items"`
	Next  string `json:"next"`
	Total int    `json:"total"`
}

type spotifyPlaylistTrackItem struct {
	Track *SpotifyTrack `json:"track"`
}

// GetSpotifyTrack fetches Spotify track metadata.
func GetSpotifyTrack(id string) (SpotifyTrack, error) {
	var track SpotifyTrack
	err := spotifyGet("tracks/"+id, &track)
	return track, err
}

// SpotifyTrackToDeezer converts a Spotify track to a Deezer track.
func SpotifyTrackToDeezer(id string) (types.TrackType, error) {
	track, err := GetSpotifyTrack(id)
	if err != nil {
		return types.TrackType{}, err
	}
	return spotifyTrackToDeezerTrack(track)
}

// GetSpotifyAlbum fetches Spotify album metadata.
func GetSpotifyAlbum(id string) (SpotifyAlbum, error) {
	var album SpotifyAlbum
	err := spotifyGet("albums/"+id, &album)
	return album, err
}

// SpotifyAlbumToDeezer converts a Spotify album to a Deezer album and track list via UPC.
func SpotifyAlbumToDeezer(id string) (types.AlbumType, []types.TrackType, error) {
	album, err := GetSpotifyAlbum(id)
	if err != nil {
		return types.AlbumType{}, nil, err
	}
	return UPCToDeezer(album.Name, album.ExternalIDs["upc"])
}

// GetSpotifyPlaylist fetches Spotify playlist metadata.
func GetSpotifyPlaylist(id string) (SpotifyPlaylist, error) {
	var playlist SpotifyPlaylist
	err := spotifyGet("playlists/"+id+"?fields=id,name,description,public,collaborative,images,owner,tracks.total,type,uri", &playlist)
	return playlist, err
}

// GetSpotifyPlaylistTracks fetches all public Spotify playlist tracks.
func GetSpotifyPlaylistTracks(id string) ([]SpotifyTrack, int, error) {
	path := "playlists/" + id + "/tracks?limit=100&fields=total,next,items(track(id,name,duration_ms,explicit,external_ids,artists(id,name,type),album(id,name,album_type,release_date,images,artists(id,name,type)),type,uri))"
	items, total, err := spotifyGetPages[spotifyPlaylistTrackItem](path)
	if err != nil {
		return nil, 0, err
	}

	tracks := make([]SpotifyTrack, 0, len(items))
	for _, item := range items {
		if item.Track == nil || item.Track.ID == "" {
			continue
		}
		tracks = append(tracks, *item.Track)
	}
	return tracks, total, nil
}

// SpotifyPlaylistToDeezer converts a Spotify playlist to Deezer playlist metadata and matching Deezer tracks.
func SpotifyPlaylistToDeezer(id string) (types.PlaylistInfo, []types.TrackType, error) {
	body, err := GetSpotifyPlaylist(id)
	var items []SpotifyTrack
	var total int
	var trackErr error
	if err == nil {
		items, total, trackErr = GetSpotifyPlaylistTracks(id)
	}
	if err != nil || trackErr != nil {
		playlist, partnerItems, partnerErr := GetSpotifyPartnerPlaylist(id)
		if partnerErr != nil {
			if err != nil {
				return types.PlaylistInfo{}, nil, fmt.Errorf("%w; spotify partner fallback failed: %v", err, partnerErr)
			}
			return types.PlaylistInfo{}, nil, fmt.Errorf("%w; spotify partner fallback failed: %v", trackErr, partnerErr)
		}
		return spotifyPlaylistInfoToDeezer(playlist, playlist.Tracks.Total), spotifyPlaylistTracksToDeezer(partnerItems), nil
	}

	return spotifyPlaylistInfoToDeezer(body, total), spotifyPlaylistTracksToDeezer(items), nil
}

func spotifyPlaylistInfoToDeezer(body SpotifyPlaylist, total int) types.PlaylistInfo {
	playlist := types.PlaylistInfo{
		PlaylistID:      body.ID,
		Description:     body.Description,
		ParentUsername:  body.Owner.DisplayName,
		ParentUserID:    body.Owner.ID,
		PictureType:     "cover",
		Title:           body.Name,
		Type:            "0",
		Status:          0,
		UserID:          body.Owner.ID,
		NbSong:          total,
		NbFan:           0,
		HasArtistLinked: false,
		IsSponsored:     false,
		IsEdito:         false,
		TYPE_INTERNAL:   "playlist",
	}
	if len(body.Images) > 0 {
		playlist.PlaylistPicture = body.Images[0].URL
	}
	return playlist
}

// GetSpotifyArtistTopTracks fetches Spotify artist top tracks.
func GetSpotifyArtistTopTracks(id string) ([]SpotifyTrack, error) {
	var result struct {
		Tracks []SpotifyTrack `json:"tracks"`
	}
	err := spotifyGet("artists/"+id+"/top-tracks?market=US", &result)
	return result.Tracks, err
}

// SpotifyArtistToDeezer converts Spotify artist top tracks to Deezer tracks.
func SpotifyArtistToDeezer(id string) ([]types.TrackType, error) {
	items, err := GetSpotifyArtistTopTracks(id)
	if err != nil {
		return nil, err
	}

	return convertTracksConcurrently(items, func(_ int, item SpotifyTrack) (types.TrackType, bool) {
		track, err := spotifyTrackToDeezerTrack(item)
		if err != nil {
			return types.TrackType{}, false
		}
		return track, true
	}), nil
}

func spotifyTrackToDeezerTrack(track SpotifyTrack) (types.TrackType, error) {
	if track.ExternalIDs["isrc"] != "" {
		result, err := ISRCToDeezer(track.Name, track.ExternalIDs["isrc"])
		if err == nil {
			return result, nil
		}
	}
	return SpotifyTrackMetadataToDeezer(track)
}

func spotifyPlaylistTracksToDeezer(items []SpotifyTrack) []types.TrackType {
	return convertTracksConcurrently(items, func(index int, item SpotifyTrack) (types.TrackType, bool) {
		track, err := spotifyTrackToDeezerTrack(item)
		if err != nil {
			return types.TrackType{}, false
		}
		position := index + 1
		track.TRACK_POSITION = &position
		return track, true
	})
}

func spotifyGet(path string, target any) error {
	var lastErr error
	for attempt := range 3 {
		status, retry, err := spotifyGetOnce(path, target)
		if err == nil {
			return nil
		}
		lastErr = err
		if status == http.StatusUnauthorized && attempt == 0 {
			resetSpotifyToken()
			continue
		}
		if !retry {
			return err
		}
		time.Sleep(time.Duration(attempt+1) * 500 * time.Millisecond)
	}
	return lastErr
}

func spotifyGetOnce(path string, target any) (int, bool, error) {
	resourceType, resourceID, err := spotifyTokenResource(path)
	if err != nil {
		return 0, false, err
	}
	token, err := getSpotifyAnonymousToken(resourceType, resourceID)
	if err != nil {
		return 0, false, err
	}

	req, err := http.NewRequest(http.MethodGet, spotifyAPIBaseURL+path, nil)
	if err != nil {
		return 0, false, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("User-Agent", spotifyBrowserUserAgent)

	resp, err := spotifyHTTPClient.Do(req)
	if err != nil {
		return 0, true, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		retryAfter := resp.Header.Get("Retry-After")
		if resp.StatusCode == http.StatusTooManyRequests && retryAfter != "" {
			return resp.StatusCode, false, fmt.Errorf("spotify API rate limited: retry after %s seconds", retryAfter)
		}
		var apiErr struct {
			Error struct {
				Status  int    `json:"status"`
				Message string `json:"message"`
			} `json:"error"`
		}
		retry := resp.StatusCode == http.StatusBadGateway ||
			resp.StatusCode == http.StatusServiceUnavailable ||
			resp.StatusCode == http.StatusGatewayTimeout
		if err := json.NewDecoder(resp.Body).Decode(&apiErr); err == nil && apiErr.Error.Message != "" {
			return resp.StatusCode, retry, fmt.Errorf("spotify API error: %s", apiErr.Error.Message)
		}
		return resp.StatusCode, retry, fmt.Errorf("spotify API error: %s", resp.Status)
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return resp.StatusCode, false, err
	}
	return resp.StatusCode, false, nil
}

func spotifyGetPages[T any](path string) ([]T, int, error) {
	items := []T{}
	total := 0
	nextURL := spotifyAPIBaseURL + path
	for nextURL != "" {
		var page spotifyList[T]
		apiPath := strings.TrimPrefix(nextURL, spotifyAPIBaseURL)
		if err := spotifyGet(apiPath, &page); err != nil {
			return nil, 0, err
		}
		if total == 0 {
			total = page.Total
		}
		items = append(items, page.Items...)
		nextURL = page.Next
	}
	return items, total, nil
}

func getSpotifyAnonymousToken(resourceType, id string) (string, error) {
	spotifyTokenMu.Lock()
	defer spotifyTokenMu.Unlock()

	key := resourceType + ":" + id
	if spotifyToken != "" && spotifyTokenKey == key && time.Now().Before(spotifyTokenExpiry.Add(-1*time.Minute)) {
		return spotifyToken, nil
	}

	embedURL := fmt.Sprintf("https://open.spotify.com/embed/%s/%s?utm_source=oembed", url.PathEscape(resourceType), url.PathEscape(id))
	req, err := http.NewRequest(http.MethodGet, embedURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("User-Agent", spotifyBrowserUserAgent)

	resp, err := spotifyHTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("spotify embed error: %s", resp.Status)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	token, expiry := extractSpotifyToken(string(bodyBytes))
	if token == "" {
		return "", fmt.Errorf("spotify access token not found")
	}

	spotifyToken = token
	spotifyTokenKey = key
	if expiry > 0 {
		spotifyTokenExpiry = time.UnixMilli(expiry)
	} else {
		spotifyTokenExpiry = time.Now().Add(30 * time.Minute)
	}
	return spotifyToken, nil
}

func resetSpotifyToken() {
	spotifyTokenMu.Lock()
	defer spotifyTokenMu.Unlock()
	spotifyToken = ""
	spotifyTokenKey = ""
	spotifyTokenExpiry = time.Time{}
}

func spotifyTokenResource(path string) (string, string, error) {
	path = strings.TrimPrefix(path, spotifyAPIBaseURL)
	path = strings.SplitN(path, "?", 2)[0]
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 2 || parts[1] == "" {
		return "", "", fmt.Errorf("unsupported spotify API path: %s", path)
	}
	switch parts[0] {
	case "tracks":
		return "track", parts[1], nil
	case "albums":
		return "album", parts[1], nil
	case "playlists":
		return "playlist", parts[1], nil
	case "artists":
		return "artist", parts[1], nil
	default:
		return "", "", fmt.Errorf("unsupported spotify API path: %s", path)
	}
}

func extractSpotifyToken(body string) (string, int64) {
	tokenMatch := regexp.MustCompile(`"accessToken":"([^"]+)"`).FindStringSubmatch(body)
	if len(tokenMatch) < 2 {
		return "", 0
	}
	expiryMatch := regexp.MustCompile(`"accessTokenExpirationTimestampMs":(\d+)`).FindStringSubmatch(body)
	if len(expiryMatch) < 2 {
		return tokenMatch[1], 0
	}
	expiry, _ := strconv.ParseInt(expiryMatch[1], 10, 64)
	return tokenMatch[1], expiry
}
