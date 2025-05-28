// internal/utils/urlparser.go
package utils

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// ParsedURLType defines the type of Spotify content.
type ParsedURLType string

const (
	SpotifyTrack    ParsedURLType = "spotify:track"
	SpotifyAlbum    ParsedURLType = "spotify:album"
	SpotifyPlaylist ParsedURLType = "spotify:playlist"
	DeezerTrack     ParsedURLType = "deezer:track"
	DeezerAlbum     ParsedURLType = "deezer:album"
	DeezerPlaylist  ParsedURLType = "deezer:playlist"
	UnknownURL      ParsedURLType = "unknown"
)

// ParsedURLInfo holds the parsed information from a music URL.
type ParsedURLInfo struct {
	Source ParsedURLType // e.g., "spotify", "deezer", "tidal"
	Type   ParsedURLType // e.g., SpotifyTrack, SpotifyAlbum
	ID     string        // The unique identifier for the track, album, or playlist
	URL    string
}

var (
	// Regex for open.spotify.com URLs
	spotifyURLRegex = regexp.MustCompile(`https?://open\.spotify\.com/(intl-\w+/)?(track|album|playlist)/([a-zA-Z0-9]+)`)
	// Regex for spotify: URI scheme
	spotifyURIRegex = regexp.MustCompile(`spotify:(track|album|playlist):([a-zA-Z0-9]+)`)
	// Regex for deezer.com URLs
	deezerURLRegex = regexp.MustCompile(`https?://(?:www\.)?deezer\.com/(?:\w+/)?(track|album|playlist)/(\d+)`)
)

// ParseMusicURL analyzes a string to determine if it's a supported music URL
// and extracts the service, type, and ID.
func ParseMusicURL(inputURL string) (*ParsedURLInfo, error) {
	inputURL = strings.TrimSpace(inputURL)
	if inputURL == "" {
		return nil, fmt.Errorf("input URL cannot be empty")
	}

	// Check Spotify URLs
	if matches := spotifyURLRegex.FindStringSubmatch(inputURL); len(matches) == 4 {
		id := matches[3]
		urlType := matches[2]
		var parsedType ParsedURLType
		switch urlType {
		case "track":
			parsedType = SpotifyTrack
		case "album":
			parsedType = SpotifyAlbum
		case "playlist":
			parsedType = SpotifyPlaylist
		default:
			return nil, fmt.Errorf("unknown spotify URL type: %s", urlType)
		}
		return &ParsedURLInfo{Source: "spotify", Type: parsedType, ID: id, URL: inputURL}, nil
	}

	// Check Spotify URIs
	if matches := spotifyURIRegex.FindStringSubmatch(inputURL); len(matches) == 3 {
		id := matches[2]
		urlType := matches[1]
		var parsedType ParsedURLType
		switch urlType {
		case "track":
			parsedType = SpotifyTrack
		case "album":
			parsedType = SpotifyAlbum
		case "playlist":
			parsedType = SpotifyPlaylist
		default:
			return nil, fmt.Errorf("unknown spotify URI type: %s", urlType)
		}
		return &ParsedURLInfo{Source: "spotify", Type: parsedType, ID: id, URL: inputURL}, nil
	}

	// Check Deezer URLs
	if matches := deezerURLRegex.FindStringSubmatch(inputURL); len(matches) == 3 {
		id := matches[2]
		urlType := matches[1]
		var parsedType ParsedURLType
		switch urlType {
		case "track":
			parsedType = DeezerTrack
		case "album":
			parsedType = DeezerAlbum
		case "playlist":
			parsedType = DeezerPlaylist
		default:
			return nil, fmt.Errorf("unknown deezer URL type: %s", urlType)
		}
		return &ParsedURLInfo{Source: "deezer", Type: parsedType, ID: id, URL: inputURL}, nil
	}

	// Attempt generic URL parsing if specific patterns fail
	parsed, err := url.Parse(inputURL)
	if err == nil && parsed.Scheme != "" && parsed.Host != "" {
		// Could potentially add more generic checks here based on hostname
		if strings.Contains(parsed.Host, "spotify.com") {
			// Generic Spotify URL detected, but couldn't determine type/ID
			return nil, fmt.Errorf("detected Spotify URL, but could not extract track/album/playlist ID: %s", inputURL)
		}
		if strings.Contains(parsed.Host, "deezer.com") {
			// Generic Deezer URL detected, but couldn't determine type/ID
			return nil, fmt.Errorf("detected Deezer URL, but could not extract track/album/playlist ID: %s", inputURL)
		}
	}

	return nil, fmt.Errorf("unsupported or unrecognized music URL format: %s", inputURL)
} 