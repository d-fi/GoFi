package converter

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/d-fi/GoFi/api"
	"github.com/d-fi/GoFi/request"
	"github.com/d-fi/GoFi/types"
)

// ISRCToDeezer finds a Deezer track by ISRC and returns the full Deezer track data.
func ISRCToDeezer(name, isrc string) (types.TrackType, error) {
	var result types.TrackType
	if isrc == "" {
		return result, fmt.Errorf("ISRC code not found for %s", name)
	}

	data, err := request.RequestPublicApi("/track/isrc:" + isrc)
	if err != nil {
		return result, fmt.Errorf("no match on deezer for %s (ISRC: %s)", name, isrc)
	}

	var publicTrack struct {
		ID int `json:"id"`
	}
	if err := json.Unmarshal(data, &publicTrack); err != nil {
		return result, err
	}
	if publicTrack.ID == 0 {
		return result, fmt.Errorf("no match on deezer for %s (ISRC: %s)", name, isrc)
	}

	return api.GetTrackInfo(fmt.Sprintf("%d", publicTrack.ID))
}

// UPCToDeezer finds a Deezer album by UPC and returns the full album data and tracks.
func UPCToDeezer(name, upc string) (types.AlbumType, []types.TrackType, error) {
	var album types.AlbumType
	if upc == "" {
		return album, nil, fmt.Errorf("UPC code not found for %s", name)
	}
	if len(upc) > 12 && strings.HasPrefix(upc, "0") {
		upc = upc[len(upc)-12:]
	}

	data, err := request.RequestPublicApi("/album/upc:" + upc)
	if err != nil {
		return album, nil, fmt.Errorf("no match on deezer for %s (UPC: %s)", name, upc)
	}

	var publicAlbum struct {
		ID int `json:"id"`
	}
	if err := json.Unmarshal(data, &publicAlbum); err != nil {
		return album, nil, err
	}
	if publicAlbum.ID == 0 {
		return album, nil, fmt.Errorf("no match on deezer for %s (UPC: %s)", name, upc)
	}

	album, err = api.GetAlbumInfo(fmt.Sprintf("%d", publicAlbum.ID))
	if err != nil {
		return album, nil, err
	}
	tracks, err := api.GetAlbumTracks(fmt.Sprintf("%d", publicAlbum.ID))
	if err != nil {
		return album, nil, err
	}
	return album, tracks.Data, nil
}
