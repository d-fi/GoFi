package api

import (
	"encoding/json"
	"fmt"

	"github.com/d-fi/GoFi/request"
	"github.com/d-fi/GoFi/types"
)

func GetTrackInfoPublicApi(sngID string) (types.TrackTypePublicAPI, error) {
	var result types.TrackTypePublicAPI
	data, err := request.RequestPublicApi("/track/" + sngID)
	if err != nil {
		return result, err
	}
	err = json.Unmarshal(data, &result)
	return result, err
}

func GetAlbumInfoPublicApi(albID string) (types.AlbumTypePublicApi, error) {
	var result types.AlbumTypePublicApi
	data, err := request.RequestPublicApi("/album/" + albID)
	if err != nil {
		return result, err
	}
	err = json.Unmarshal(data, &result)
	return result, err
}

func GetTrackInfo(sngID string) (types.TrackType, error) {
	var result types.TrackType
	data, err := request.Request(map[string]interface{}{"sng_id": sngID}, "song.getData")
	if err != nil {
		return result, err
	}
	err = json.Unmarshal(data, &result)
	return result, err
}

func GetLyrics(sngID string) (types.LyricsType, error) {
	var result types.LyricsType
	data, err := request.Request(map[string]interface{}{"sng_id": sngID}, "song.getLyrics")
	if err != nil {
		return result, err
	}
	err = json.Unmarshal(data, &result)
	return result, err
}

func GetAlbumInfo(albID string) (types.AlbumType, error) {
	var result types.AlbumType
	data, err := request.Request(map[string]interface{}{"alb_id": albID}, "album.getData")
	if err != nil {
		return result, err
	}
	err = json.Unmarshal(data, &result)
	return result, err
}

func GetAlbumTracks(albID string) (types.AlbumTracksType, error) {
	var result types.AlbumTracksType
	data, err := request.Request(map[string]interface{}{
		"alb_id": albID,
		"lang":   "us",
		"nb":     -1,
	}, "song.getListByAlbum")
	if err != nil {
		return result, err
	}
	err = json.Unmarshal(data, &result)
	return result, err
}

func GetPlaylistInfo(playlistID string) (types.PlaylistInfo, error) {
	var result types.PlaylistInfo
	data, err := request.Request(map[string]interface{}{
		"playlist_id": playlistID,
		"lang":        "en",
	}, "playlist.getData")
	if err != nil {
		return result, err
	}
	err = json.Unmarshal(data, &result)
	return result, err
}

func GetPlaylistTracks(playlistID string) (types.PlaylistTracksType, error) {
	var result types.PlaylistTracksType
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
		return result, err
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return result, fmt.Errorf("failed to unmarshal playlist tracks data: %v", err)
	}

	// Update track position in the playlist
	for index, track := range result.Data {
		// Check if TrackPosition is a pointer and allocate a new int if necessary
		if track.TRACK_POSITION == nil {
			track.TRACK_POSITION = new(int)
		}
		*track.TRACK_POSITION = index + 1
	}

	return result, nil
}

func GetArtistInfo(artID string) (types.ArtistInfoType, error) {
	var result types.ArtistInfoType
	data, err := request.Request(map[string]interface{}{
		"art_id":         artID,
		"filter_role_id": []int{0},
		"lang":           "en",
		"tab":            0,
		"nb":             -1,
		"start":          0,
	}, "artist.getData")
	if err != nil {
		return result, err
	}
	err = json.Unmarshal(data, &result)
	return result, err
}

func GetDiscography(artID string, nb int) (types.DiscographyType, error) {
	var result types.DiscographyType
	data, err := request.Request(map[string]interface{}{
		"art_id":         artID,
		"filter_role_id": []int{0},
		"lang":           "en",
		"nb":             nb,
		"nb_songs":       -1,
		"start":          0,
	}, "album.getDiscography")
	if err != nil {
		return result, err
	}
	err = json.Unmarshal(data, &result)
	return result, err
}

func GetProfile(userID string) (types.ProfileType, error) {
	var result types.ProfileType
	data, err := request.Request(map[string]interface{}{
		"user_id": userID,
		"tab":     "loved",
		"nb":      -1,
	}, "mobile.pageUser")
	if err != nil {
		return result, err
	}
	err = json.Unmarshal(data, &result)
	return result, err
}

func SearchAlternative(artist, song string, nb int) (types.SearchType, error) {
	var result types.SearchType
	data, err := request.Request(map[string]interface{}{
		"query": fmt.Sprintf("artist:'%s' track:'%s'", artist, song),
		"types": []string{"TRACK"},
		"nb":    nb,
	}, "mobile_suggest")
	if err != nil {
		return result, err
	}
	err = json.Unmarshal(data, &result)
	return result, err
}

func SearchMusic(query string, nb int, searchTypes ...string) (types.SearchType, error) {
	var result types.SearchType

	// Set default types if none are provided
	if len(searchTypes) == 0 {
		searchTypes = []string{"TRACK"}
	}

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
		return result, err
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return result, fmt.Errorf("failed to unmarshal search results: %v", err)
	}

	return result, nil
}

func GetUser() (types.UserType, error) {
	var result types.UserType
	data, err := request.RequestGet("user_getInfo", nil)
	if err != nil {
		return result, err
	}
	err = json.Unmarshal(data, &result)
	return result, err
}

func GetChannelList() (types.ChannelSearchType, error) {
	var result types.ChannelSearchType
	data, err := request.Request(nil, "search_getChannels")
	if err != nil {
		return result, err
	}
	err = json.Unmarshal(data, &result)
	return result, err
}

func GetShowInfo(showID string, nb, start int) (types.ShowType, error) {
	var result types.ShowType
	data, err := request.Request(map[string]interface{}{
		"SHOW_ID": showID,
		"NB":      nb,
		"START":   start,
	}, "mobile.pageShow")
	if err != nil {
		return result, err
	}
	err = json.Unmarshal(data, &result)
	return result, err
}
