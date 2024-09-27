package api

import (
	"encoding/json"
	"fmt"

	"github.com/d-fi/GoFi/logger"
	"github.com/d-fi/GoFi/request"
	"github.com/d-fi/GoFi/types"
)

// GetTrackInfoPublicApi fetches public track information from the API.
func GetTrackInfoPublicApi(sngID string) (types.TrackTypePublicAPI, error) {
	var result types.TrackTypePublicAPI
	logger.Debug("Requesting track info from public API for ID: %s", sngID)
	data, err := request.RequestPublicApi("/track/" + sngID)
	if err != nil {
		logger.Error("Failed to fetch track info: %v", err)
		return result, err
	}
	err = json.Unmarshal(data, &result)
	if err != nil {
		logger.Error("Failed to unmarshal track info: %v", err)
	}
	return result, err
}

// GetAlbumInfoPublicApi fetches public album information from the API.
func GetAlbumInfoPublicApi(albID string) (types.AlbumTypePublicApi, error) {
	var result types.AlbumTypePublicApi
	logger.Debug("Requesting album info from public API for ID: %s", albID)
	data, err := request.RequestPublicApi("/album/" + albID)
	if err != nil {
		logger.Error("Failed to fetch album info: %v", err)
		return result, err
	}
	err = json.Unmarshal(data, &result)
	if err != nil {
		logger.Error("Failed to unmarshal album info: %v", err)
	}
	return result, err
}

// GetTrackInfo fetches detailed track information.
func GetTrackInfo(sngID string) (types.TrackType, error) {
	var result types.TrackType
	logger.Debug("Requesting detailed track info for ID: %s", sngID)
	data, err := request.Request(map[string]interface{}{"sng_id": sngID}, "song.getData")
	if err != nil {
		logger.Error("Failed to fetch detailed track info: %v", err)
		return result, err
	}
	err = json.Unmarshal(data, &result)
	if err != nil {
		logger.Error("Failed to unmarshal detailed track info: %v", err)
	}
	return result, err
}

// GetLyrics fetches lyrics for a given track.
func GetLyrics(sngID string) (types.LyricsType, error) {
	var result types.LyricsType
	logger.Debug("Requesting lyrics for track ID: %s", sngID)
	data, err := request.Request(map[string]interface{}{"sng_id": sngID}, "song.getLyrics")
	if err != nil {
		logger.Error("Failed to fetch lyrics: %v", err)
		return result, err
	}
	err = json.Unmarshal(data, &result)
	if err != nil {
		logger.Error("Failed to unmarshal lyrics: %v", err)
	}
	return result, err
}

// GetAlbumInfo fetches detailed album information.
func GetAlbumInfo(albID string) (types.AlbumType, error) {
	var result types.AlbumType
	logger.Debug("Requesting detailed album info for ID: %s", albID)
	data, err := request.Request(map[string]interface{}{"alb_id": albID}, "album.getData")
	if err != nil {
		logger.Error("Failed to fetch detailed album info: %v", err)
		return result, err
	}
	err = json.Unmarshal(data, &result)
	if err != nil {
		logger.Error("Failed to unmarshal detailed album info: %v", err)
	}
	return result, err
}

// GetAlbumTracks fetches tracks of a given album.
func GetAlbumTracks(albID string) (types.AlbumTracksType, error) {
	var result types.AlbumTracksType
	logger.Debug("Requesting tracks for album ID: %s", albID)
	data, err := request.Request(map[string]interface{}{
		"alb_id": albID,
		"lang":   "us",
		"nb":     -1,
	}, "song.getListByAlbum")
	if err != nil {
		logger.Error("Failed to fetch album tracks: %v", err)
		return result, err
	}
	err = json.Unmarshal(data, &result)
	if err != nil {
		logger.Error("Failed to unmarshal album tracks: %v", err)
	}
	return result, err
}

// GetPlaylistInfo fetches information about a playlist.
func GetPlaylistInfo(playlistID string) (types.PlaylistInfo, error) {
	var result types.PlaylistInfo
	logger.Debug("Requesting playlist info for ID: %s", playlistID)
	data, err := request.Request(map[string]interface{}{
		"playlist_id": playlistID,
		"lang":        "en",
	}, "playlist.getData")
	if err != nil {
		logger.Error("Failed to fetch playlist info: %v", err)
		return result, err
	}
	err = json.Unmarshal(data, &result)
	if err != nil {
		logger.Error("Failed to unmarshal playlist info: %v", err)
	}
	return result, err
}

// GetPlaylistTracks fetches tracks in a given playlist.
func GetPlaylistTracks(playlistID string) (types.PlaylistTracksType, error) {
	var result types.PlaylistTracksType
	logger.Debug("Requesting playlist tracks for ID: %s", playlistID)
	data, err := request.Request(map[string]interface{}{
		"playlist_id": playlistID,
		"lang":        "en",
		"nb":          -1,
		"start":       0,
		"tab":         0,
		"tags":        true,
		"header":      true,
	}, "playlist.getSongs")
	if err != nil {
		logger.Error("Failed to fetch playlist tracks: %v", err)
		return result, err
	}

	if err := json.Unmarshal(data, &result); err != nil {
		logger.Error("Failed to unmarshal playlist tracks data: %v", err)
		return result, fmt.Errorf("failed to unmarshal playlist tracks data: %v", err)
	}

	for index, track := range result.Data {
		if track.TRACK_POSITION == nil {
			track.TRACK_POSITION = new(int)
		}
		*track.TRACK_POSITION = index + 1
	}

	return result, nil
}

// GetArtistInfo fetches information about an artist.
func GetArtistInfo(artID string) (types.ArtistInfoType, error) {
	var result types.ArtistInfoType
	logger.Debug("Requesting artist info for ID: %s", artID)
	data, err := request.Request(map[string]interface{}{
		"art_id":         artID,
		"filter_role_id": []int{0},
		"lang":           "en",
		"tab":            0,
		"nb":             -1,
		"start":          0,
	}, "artist.getData")
	if err != nil {
		logger.Error("Failed to fetch artist info: %v", err)
		return result, err
	}
	err = json.Unmarshal(data, &result)
	if err != nil {
		logger.Error("Failed to unmarshal artist info: %v", err)
	}
	return result, err
}

// GetDiscography fetches an artist's discography.
func GetDiscography(artID string, nb int) (types.DiscographyType, error) {
	var result types.DiscographyType
	logger.Debug("Requesting discography for artist ID: %s", artID)
	data, err := request.Request(map[string]interface{}{
		"art_id":         artID,
		"filter_role_id": []int{0},
		"lang":           "en",
		"nb":             nb,
		"nb_songs":       -1,
		"start":          0,
	}, "album.getDiscography")
	if err != nil {
		logger.Error("Failed to fetch discography: %v", err)
		return result, err
	}
	err = json.Unmarshal(data, &result)
	if err != nil {
		logger.Error("Failed to unmarshal discography: %v", err)
	}
	return result, err
}

// GetProfile fetches user profile information.
func GetProfile(userID string) (types.ProfileType, error) {
	var result types.ProfileType
	logger.Debug("Requesting profile info for user ID: %s", userID)
	data, err := request.Request(map[string]interface{}{
		"user_id": userID,
		"tab":     "loved",
		"nb":      -1,
	}, "mobile.pageUser")
	if err != nil {
		logger.Error("Failed to fetch profile info: %v", err)
		return result, err
	}
	err = json.Unmarshal(data, &result)
	if err != nil {
		logger.Error("Failed to unmarshal profile info: %v", err)
	}
	return result, err
}

// SearchAlternative searches for alternative tracks by artist and song name.
func SearchAlternative(artist, song string, nb int) (types.SearchType, error) {
	var result types.SearchType
	logger.Debug("Searching for alternative tracks by artist: %s and song: %s", artist, song)
	data, err := request.Request(map[string]interface{}{
		"query": fmt.Sprintf("artist:'%s' track:'%s'", artist, song),
		"types": []string{"TRACK"},
		"nb":    nb,
	}, "mobile_suggest")
	if err != nil {
		logger.Error("Failed to search for alternative tracks: %v", err)
		return result, err
	}
	err = json.Unmarshal(data, &result)
	if err != nil {
		logger.Error("Failed to unmarshal search results: %v", err)
	}
	return result, err
}

// SearchMusic searches for music based on a query.
func SearchMusic(query string, nb int, searchTypes ...string) (types.SearchType, error) {
	var result types.SearchType

	if len(searchTypes) == 0 {
		searchTypes = []string{"TRACK"}
	}

	logger.Debug("Searching music with query: %s", query)
	data, err := request.Request(map[string]interface{}{
		"query":          query,
		"start":          0,
		"nb":             nb,
		"types":          searchTypes,
		"suggest":        true,
		"artist_suggest": true,
		"top_tracks":     true,
	}, "mobile_suggest")

	if err != nil {
		logger.Error("Failed to search music: %v", err)
		return result, err
	}

	if err := json.Unmarshal(data, &result); err != nil {
		logger.Error("Failed to unmarshal search results: %v", err)
		return result, fmt.Errorf("failed to unmarshal search results: %v", err)
	}

	return result, nil
}

// GetUser fetches the current user's information.
func GetUser() (types.UserType, error) {
	var result types.UserType
	logger.Debug("Fetching current user info")
	data, err := request.RequestGet("user_getInfo", nil)
	if err != nil {
		logger.Error("Failed to fetch user info: %v", err)
		return result, err
	}
	err = json.Unmarshal(data, &result)
	if err != nil {
		logger.Error("Failed to unmarshal user info: %v", err)
	}
	return result, err
}

// GetChannelList fetches a list of available channels.
func GetChannelList() (types.ChannelSearchType, error) {
	var result types.ChannelSearchType
	logger.Debug("Fetching channel list")
	data, err := request.Request(nil, "search_getChannels")
	if err != nil {
		logger.Error("Failed to fetch channel list: %v", err)
		return result, err
	}
	err = json.Unmarshal(data, &result)
	if err != nil {
		logger.Error("Failed to unmarshal channel list: %v", err)
	}
	return result, err
}

// GetShowInfo fetches information about a show.
func GetShowInfo(showID string, nb, start int) (types.ShowType, error) {
	var result types.ShowType
	logger.Debug("Fetching show info for ID: %s", showID)
	data, err := request.Request(map[string]interface{}{
		"SHOW_ID": showID,
		"NB":      nb,
		"START":   start,
	}, "mobile.pageShow")
	if err != nil {
		logger.Error("Failed to fetch show info: %v", err)
		return result, err
	}
	err = json.Unmarshal(data, &result)
	if err != nil {
		logger.Error("Failed to unmarshal show info: %v", err)
	}
	return result, err
}
