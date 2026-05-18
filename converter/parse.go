package converter

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/d-fi/GoFi/api"
	"github.com/d-fi/GoFi/types"
)

type URLParts struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type ParseResult struct {
	Info     URLParts          `json:"info"`
	LinkType string            `json:"linktype"`
	LinkInfo interface{}       `json:"linkinfo"`
	Tracks   []types.TrackType `json:"tracks"`
}

// GetURLParts parses supported Deezer and YouTube URLs into an id/type pair.
func GetURLParts(rawURL string) (URLParts, error) {
	if strings.Contains(rawURL, "spotify") || strings.HasPrefix(rawURL, "spotify:") {
		return URLParts{}, fmt.Errorf("spotify URLs are not supported: %s", rawURL)
	}
	if strings.Contains(rawURL, "tidal") {
		return URLParts{}, fmt.Errorf("tidal URLs are not supported: %s", rawURL)
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return URLParts{}, err
	}

	host := strings.ToLower(parsed.Host)
	switch {
	case strings.Contains(host, "deezer"):
		if strings.Contains(host, "page.link") {
			rawURL, err = resolveRedirect(rawURL)
			if err != nil {
				return URLParts{}, err
			}
		}
		return parseDeezerURL(rawURL)
	case strings.Contains(host, "youtube.com"):
		id := parsed.Query().Get("v")
		if id == "" {
			return URLParts{}, fmt.Errorf("unable to parse id")
		}
		return URLParts{Type: "youtube-track", ID: id}, nil
	case strings.Contains(host, "youtu.be"):
		id := strings.Trim(strings.TrimPrefix(parsed.Path, "/"), "/")
		if id == "" {
			return URLParts{}, fmt.Errorf("unable to parse id")
		}
		return URLParts{Type: "youtube-track", ID: id}, nil
	default:
		return URLParts{}, fmt.Errorf("unknown URL: %s", rawURL)
	}
}

// ParseInfo resolves a supported Deezer or YouTube URL into Deezer tracks.
func ParseInfo(rawURL string) (ParseResult, error) {
	info, err := GetURLParts(rawURL)
	if err != nil {
		return ParseResult{}, err
	}
	if info.ID == "" {
		return ParseResult{}, fmt.Errorf("unable to parse id")
	}

	result := ParseResult{
		Info:     info,
		LinkType: "track",
		LinkInfo: map[string]interface{}{},
		Tracks:   []types.TrackType{},
	}

	switch info.Type {
	case "track":
		track, err := api.GetTrackInfo(info.ID)
		if err != nil {
			return result, err
		}
		result.Tracks = append(result.Tracks, track)
	case "album", "audiobook":
		album, err := api.GetAlbumInfo(info.ID)
		if err != nil {
			return result, err
		}
		tracks, err := api.GetAlbumTracks(info.ID)
		if err != nil {
			return result, err
		}
		result.LinkType = "album"
		result.LinkInfo = album
		result.Tracks = tracks.Data
	case "playlist":
		playlist, err := api.GetPlaylistInfo(info.ID)
		if err != nil {
			return result, err
		}
		tracks, err := api.GetPlaylistTracks(info.ID)
		if err != nil {
			return result, err
		}
		result.LinkType = "playlist"
		result.LinkInfo = playlist
		result.Tracks = tracks.Data
	case "artist":
		artist, err := api.GetArtistInfo(info.ID)
		if err != nil {
			return result, err
		}
		albums, err := api.GetDiscography(info.ID, 500)
		if err != nil {
			return result, err
		}
		result.LinkType = "artist"
		result.LinkInfo = artist
		for _, album := range albums.Data {
			if !albumContainsArtist(album, info.ID) {
				continue
			}
			albumTracks, err := api.GetAlbumTracks(album.ALB_ID)
			if err != nil {
				return result, err
			}
			for _, track := range albumTracks.Data {
				if track.ART_ID == info.ID {
					result.Tracks = append(result.Tracks, track)
				}
			}
		}
	case "youtube-track":
		track, err := YouTubeTrackToDeezer(info.ID)
		if err != nil {
			return result, err
		}
		result.Tracks = append(result.Tracks, track)
	default:
		return result, fmt.Errorf("unknown type: %s", info.Type)
	}

	for i := range result.Tracks {
		version := result.Tracks[i].VERSION
		if version != nil && *version != "" && !strings.Contains(result.Tracks[i].SNG_TITLE, *version) {
			result.Tracks[i].SNG_TITLE += " " + *version
		}
	}

	return result, nil
}

func parseDeezerURL(rawURL string) (URLParts, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return URLParts{}, err
	}
	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	for i := 0; i < len(parts)-1; i++ {
		switch parts[i] {
		case "track", "album", "audiobook", "artist", "playlist":
			if parts[i+1] != "" {
				return URLParts{Type: parts[i], ID: parts[i+1]}, nil
			}
		}
	}

	re := regexp.MustCompile(`/(track|album|audiobook|artist|playlist)/(\d+)`)
	matches := re.FindStringSubmatch(rawURL)
	if len(matches) == 3 {
		return URLParts{Type: matches[1], ID: matches[2]}, nil
	}
	return URLParts{}, fmt.Errorf("unable to parse URL: %s", rawURL)
}

func resolveRedirect(rawURL string) (string, error) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Head(rawURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if location := resp.Header.Get("Location"); location != "" {
		redirectURL, err := resp.Request.URL.Parse(location)
		if err != nil {
			return "", err
		}
		return redirectURL.String(), nil
	}
	return rawURL, nil
}

func albumContainsArtist(album types.AlbumType, artistID string) bool {
	for _, artist := range album.ARTISTS {
		if artist.ART_ID == artistID {
			return true
		}
	}
	return false
}
