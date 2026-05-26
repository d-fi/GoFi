package converter

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/d-fi/GoFi/api"
	"github.com/d-fi/GoFi/types"
)

const spotifyAPIBaseURL = "https://api.spotify.com/v1/"

var (
	spotifyTokenMu     sync.Mutex
	spotifyToken       string
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

// SpotifyTrackToDeezer converts a Spotify track to a Deezer track via ISRC, with search fallback.
func SpotifyTrackToDeezer(id string) (types.TrackType, error) {
	track, err := GetSpotifyTrack(id)
	if err == nil {
		return spotifyTrackToDeezerTrack(track)
	}
	if fallback, fallbackErr := spotifyOEmbedTrackToDeezer(id); fallbackErr == nil {
		return fallback, nil
	}
	return types.TrackType{}, err
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
	if err == nil {
		deezerAlbum, tracks, convertErr := UPCToDeezer(album.Name, album.ExternalIDs["upc"])
		if convertErr == nil {
			return deezerAlbum, tracks, nil
		}
		err = convertErr
	}
	if fallbackAlbum, fallbackTracks, fallbackErr := spotifyOEmbedAlbumToDeezer(id); fallbackErr == nil {
		return fallbackAlbum, fallbackTracks, nil
	}
	return types.AlbumType{}, nil, err
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
	if err != nil {
		return types.PlaylistInfo{}, nil, err
	}
	items, total, err := GetSpotifyPlaylistTracks(id)
	if err != nil {
		return types.PlaylistInfo{}, nil, err
	}

	tracks := make([]types.TrackType, 0, len(items))
	for index, item := range items {
		track, err := spotifyTrackToDeezerTrack(item)
		if err != nil {
			continue
		}
		position := index + 1
		track.TRACK_POSITION = &position
		tracks = append(tracks, track)
	}

	playlist := types.PlaylistInfo{
		PlaylistID:      body.ID,
		Description:     body.Description,
		ParentUsername:  body.Owner.DisplayName,
		ParentUserID:    body.Owner.ID,
		PictureType:     "cover",
		PlaylistPicture: firstSpotifyImage(body.Images),
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
	return playlist, tracks, nil
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

	tracks := make([]types.TrackType, 0, len(items))
	for _, item := range items {
		track, err := spotifyTrackToDeezerTrack(item)
		if err != nil {
			continue
		}
		tracks = append(tracks, track)
	}
	return tracks, nil
}

func spotifyTrackToDeezerTrack(track SpotifyTrack) (types.TrackType, error) {
	if track.ExternalIDs["isrc"] != "" {
		return ISRCToDeezer(track.Name, track.ExternalIDs["isrc"])
	}
	if len(track.Artists) > 0 {
		search, err := api.SearchAlternative(track.Artists[0].Name, track.Name, 1)
		if err == nil && len(search.TRACK.Data) > 0 {
			return search.TRACK.Data[0], nil
		}
	}

	query := strings.TrimSpace(track.Name + " " + spotifyArtistsString(track.Artists))
	if query != "" {
		search, err := api.SearchMusic(query, 20, "TRACK")
		if err == nil && len(search.TRACK.Data) > 0 {
			return search.TRACK.Data[0], nil
		}
	}
	return types.TrackType{}, fmt.Errorf("no track found for spotify track %s", track.ID)
}

func spotifyGet(path string, target any) error {
	token, err := getSpotifyAnonymousToken("track", "7FIWs0pqAYbP91WWM0vlTQ")
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodGet, spotifyAPIBaseURL+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("User-Agent", spotifyUserAgent())

	resp, err := spotifyHTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		retryAfter := resp.Header.Get("Retry-After")
		if resp.StatusCode == http.StatusTooManyRequests && retryAfter != "" {
			return fmt.Errorf("spotify API rate limited: retry after %s seconds", retryAfter)
		}
		var apiErr struct {
			Error struct {
				Status  int    `json:"status"`
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&apiErr); err == nil && apiErr.Error.Message != "" {
			return fmt.Errorf("spotify API error: %s", apiErr.Error.Message)
		}
		return fmt.Errorf("spotify API error: %s", resp.Status)
	}

	return json.NewDecoder(resp.Body).Decode(target)
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

	if spotifyToken != "" && time.Now().Before(spotifyTokenExpiry.Add(-1*time.Minute)) {
		return spotifyToken, nil
	}

	embedURL := fmt.Sprintf("https://open.spotify.com/embed/%s/%s?utm_source=oembed", url.PathEscape(resourceType), url.PathEscape(id))
	req, err := http.NewRequest(http.MethodGet, embedURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("User-Agent", spotifyUserAgent())

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
	if expiry > 0 {
		spotifyTokenExpiry = time.UnixMilli(expiry)
	} else {
		spotifyTokenExpiry = time.Now().Add(30 * time.Minute)
	}
	return spotifyToken, nil
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

func spotifyOEmbedTrackToDeezer(id string) (types.TrackType, error) {
	metadata, metadataErr := fetchSpotifyEmbedMetadata("track", id)
	if metadataErr == nil {
		cleanTitle := cleanSpotifyTitle(metadata.Title)
		if metadata.Title != "" && len(metadata.Artists) > 0 {
			search, err := api.SearchAlternative(metadata.Artists[0], metadata.Title, 1)
			if err == nil && len(search.TRACK.Data) > 0 {
				return search.TRACK.Data[0], nil
			}
			if cleanTitle != "" && cleanTitle != metadata.Title {
				search, err := api.SearchAlternative(metadata.Artists[0], cleanTitle, 1)
				if err == nil && len(search.TRACK.Data) > 0 {
					return search.TRACK.Data[0], nil
				}
			}
		}
		for _, query := range []string{
			strings.TrimSpace(cleanTitle + " " + strings.Join(metadata.Artists, " ")),
			strings.TrimSpace(metadata.Title + " " + strings.Join(metadata.Artists, " ")),
		} {
			if query == "" {
				continue
			}
			search, err := api.SearchMusic(query, 20, "TRACK")
			if err == nil && len(search.TRACK.Data) > 0 {
				return search.TRACK.Data[0], nil
			}
		}
	}

	title, err := fetchSpotifyOEmbedTitle("track", id)
	if err != nil {
		return types.TrackType{}, err
	}
	search, err := api.SearchMusic(title, 20, "TRACK")
	if err != nil {
		return types.TrackType{}, err
	}
	if len(search.TRACK.Data) == 0 {
		return types.TrackType{}, fmt.Errorf("no track found for spotify track %s", id)
	}
	return search.TRACK.Data[0], nil
}

func spotifyOEmbedAlbumToDeezer(id string) (types.AlbumType, []types.TrackType, error) {
	title := ""
	if metadata, err := fetchSpotifyEmbedMetadata("album", id); err == nil {
		title = strings.TrimSpace(metadata.Title + " " + strings.Join(metadata.Artists, " "))
	}
	if title == "" {
		oembedTitle, err := fetchSpotifyOEmbedTitle("album", id)
		if err != nil {
			return types.AlbumType{}, nil, err
		}
		title = oembedTitle
	}
	search, err := api.SearchMusic(title, 10, "ALBUM")
	if err != nil {
		return types.AlbumType{}, nil, err
	}
	if len(search.ALBUM.Data) == 0 {
		return types.AlbumType{}, nil, fmt.Errorf("no album found for spotify album %s", id)
	}
	albumID := search.ALBUM.Data[0].ALB_ID
	album, err := api.GetAlbumInfo(albumID)
	if err != nil {
		return types.AlbumType{}, nil, err
	}
	tracks, err := api.GetAlbumTracks(albumID)
	if err != nil {
		return types.AlbumType{}, nil, err
	}
	return album, tracks.Data, nil
}

type spotifyEmbedMetadata struct {
	Title   string
	Artists []string
}

func fetchSpotifyEmbedMetadata(resourceType, id string) (spotifyEmbedMetadata, error) {
	var metadata spotifyEmbedMetadata
	embedURL := fmt.Sprintf("https://open.spotify.com/embed/%s/%s?utm_source=oembed", url.PathEscape(resourceType), url.PathEscape(id))
	req, err := http.NewRequest(http.MethodGet, embedURL, nil)
	if err != nil {
		return metadata, err
	}
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("User-Agent", spotifyUserAgent())

	resp, err := spotifyHTTPClient.Do(req)
	if err != nil {
		return metadata, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return metadata, fmt.Errorf("spotify embed error: %s", resp.Status)
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return metadata, err
	}
	body := string(bodyBytes)

	metadata.Title = firstJSONRegexValue(body, `"title":"([^"]+)"`)
	if metadata.Title == "" {
		metadata.Title = firstJSONRegexValue(body, `"name":"([^"]+)"`)
	}
	artistBlock := regexp.MustCompile(`"artists":\[(.*?)\]`).FindStringSubmatch(body)
	if len(artistBlock) > 1 {
		for _, match := range regexp.MustCompile(`"name":"([^"]+)"`).FindAllStringSubmatch(artistBlock[1], -1) {
			if len(match) > 1 {
				metadata.Artists = append(metadata.Artists, unquoteJSONString(match[1]))
			}
		}
	}

	if metadata.Title == "" && len(metadata.Artists) == 0 {
		return metadata, fmt.Errorf("spotify embed metadata not found")
	}
	return metadata, nil
}

func fetchSpotifyOEmbedTitle(resourceType, id string) (string, error) {
	oembedURL := fmt.Sprintf("https://open.spotify.com/oembed?url=https://open.spotify.com/%s/%s", url.PathEscape(resourceType), url.PathEscape(id))
	req, err := http.NewRequest(http.MethodGet, oembedURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", spotifyUserAgent())

	resp, err := spotifyHTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("spotify oembed error: %s", resp.Status)
	}

	var data struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}
	if data.Title == "" {
		return "", fmt.Errorf("spotify oembed title not found")
	}
	return html.UnescapeString(data.Title), nil
}

func firstJSONRegexValue(body, pattern string) string {
	match := regexp.MustCompile(pattern).FindStringSubmatch(body)
	if len(match) < 2 {
		return ""
	}
	return unquoteJSONString(match[1])
}

func unquoteJSONString(value string) string {
	unquoted, err := strconv.Unquote(`"` + value + `"`)
	if err != nil {
		return html.UnescapeString(value)
	}
	return html.UnescapeString(unquoted)
}

func cleanSpotifyTitle(title string) string {
	title = regexp.MustCompile(`(?i)\s*[\(\[]\s*(feat\.?|featuring|with)\b[^\)\]]*[\)\]]`).ReplaceAllString(title, "")
	title = regexp.MustCompile(`(?i)\s+-\s+(feat\.?|featuring|with)\b.*$`).ReplaceAllString(title, "")
	return strings.TrimSpace(title)
}

func spotifyArtistsString(artists []SpotifyArtist) string {
	names := make([]string, 0, len(artists))
	for _, artist := range artists {
		if artist.Name != "" {
			names = append(names, artist.Name)
		}
	}
	return strings.Join(names, " ")
}

func firstSpotifyImage(images []SpotifyImage) string {
	if len(images) == 0 {
		return ""
	}
	return images[0].URL
}

func spotifyUserAgent() string {
	return "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0 Safari/537.36"
}
