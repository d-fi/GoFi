package dfi

import (
	"fmt"
	"strings"

	"github.com/d-fi/GoFi/api"
	"github.com/d-fi/GoFi/converter"
	"github.com/d-fi/GoFi/types"
)

const (
	SearchOptionLimit = 50
	TrackSearchLimit  = 15
)

type ResolvedInput struct {
	Info     converter.URLParts
	LinkType string
	LinkInfo any
	Tracks   []types.TrackType
}

type SearchOption struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
}

func ParseResolvedURL(rawURL string) (ResolvedInput, error) {
	data, err := converter.ParseInfo(rawURL)
	if err != nil {
		return ResolvedInput{}, err
	}
	return ResolvedInput{
		Info:     data.Info,
		LinkType: data.LinkType,
		LinkInfo: data.LinkInfo,
		Tracks:   data.Tracks,
	}, nil
}

func ResolveTrackSearch(query string) (ResolvedInput, error) {
	search, err := api.SearchMusic(query, TrackSearchLimit, "TRACK")
	if err != nil {
		return ResolvedInput{}, err
	}
	tracks := append([]types.TrackType(nil), search.TRACK.Data...)
	tracks = AppendTrackVersionsToTitles(tracks)
	return ResolvedInput{
		Info:     converter.URLParts{Type: "track", ID: query},
		LinkType: "track",
		LinkInfo: map[string]any{},
		Tracks:   tracks,
	}, nil
}

func SearchOptions(searchType, query string, limit int) ([]SearchOption, error) {
	if query == "" {
		return nil, fmt.Errorf("missing search text")
	}
	if limit <= 0 {
		limit = SearchOptionLimit
	}
	switch searchType {
	case "artist":
		search, err := api.SearchMusic(query, limit, "ARTIST")
		if err != nil {
			return nil, err
		}
		options := make([]SearchOption, 0, len(search.ARTIST.Data))
		for _, item := range search.ARTIST.Data {
			options = append(options, SearchOption{
				Title:       item.ART_NAME,
				Description: fmt.Sprintf("%d fans", item.NB_FAN),
				URL:         "https://deezer.com/us/artist/" + item.ART_ID,
			})
		}
		return options, nil
	case "album":
		search, err := api.SearchMusic(query, limit, "ALBUM")
		if err != nil {
			return nil, err
		}
		options := make([]SearchOption, 0, len(search.ALBUM.Data))
		for _, item := range search.ALBUM.Data {
			options = append(options, SearchOption{
				Title:       item.ALB_TITLE,
				Description: fmt.Sprintf("by %s, %s tracks", item.ART_NAME, item.NUMBER_TRACK),
				URL:         "https://deezer.com/us/album/" + item.ALB_ID,
			})
		}
		return options, nil
	case "playlist":
		search, err := api.SearchMusic(query, limit, "PLAYLIST")
		if err != nil {
			return nil, err
		}
		options := make([]SearchOption, 0, len(search.PLAYLIST.Data))
		for _, item := range search.PLAYLIST.Data {
			options = append(options, SearchOption{
				Title:       item.Title,
				Description: fmt.Sprintf("by %s, %d tracks", item.ParentUsername, item.NbSong),
				URL:         "https://deezer.com/us/playlist/" + item.PlaylistID,
			})
		}
		return options, nil
	default:
		return nil, fmt.Errorf("unsupported search type: %s", searchType)
	}
}

func FirstSearchResultURL(searchType, query string) (string, error) {
	options, err := SearchOptions(searchType, query, 1)
	if err != nil {
		return "", err
	}
	if len(options) == 0 {
		return "", fmt.Errorf("no %s found", searchType)
	}
	return options[0].URL, nil
}

func SelectTracksByIndexes(tracks []types.TrackType, indexes []int) []types.TrackType {
	if len(indexes) == 0 {
		return tracks
	}
	out := make([]types.TrackType, 0, len(indexes))
	seen := map[int]bool{}
	for _, index := range indexes {
		if index < 0 || index >= len(tracks) || seen[index] {
			continue
		}
		seen[index] = true
		out = append(out, tracks[index])
	}
	return out
}

func ParseQualityStrict(value string) (quality int, ext string, label string, err error) {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "", "320", "3", "mp3_320", "320kbps", "128", "1", "mp3_128", "128kbps", "flac", "9":
		quality, ext, label = ParseQuality(value)
		return quality, ext, label, nil
	default:
		return 0, "", "", fmt.Errorf("invalid quality: %s", value)
	}
}
