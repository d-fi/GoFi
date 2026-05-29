package converter

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type spotifyPartnerPlaylistResponse struct {
	Data struct {
		PlaylistV2 struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			Description string `json:"description"`
			URI         string `json:"uri"`
			Followers   int    `json:"followers"`
			Images      struct {
				Items []struct {
					Sources []SpotifyImage `json:"sources"`
				} `json:"items"`
			} `json:"images"`
			OwnerV2 struct {
				Data struct {
					Name     string `json:"name"`
					URI      string `json:"uri"`
					Username string `json:"username"`
				} `json:"data"`
			} `json:"ownerV2"`
			Content struct {
				Items []struct {
					ItemV2 struct {
						Data spotifyPartnerTrack `json:"data"`
					} `json:"itemV2"`
				} `json:"items"`
				PagingInfo struct {
					NextOffset *int `json:"nextOffset"`
				} `json:"pagingInfo"`
				TotalCount int `json:"totalCount"`
			} `json:"content"`
		} `json:"playlistV2"`
	} `json:"data"`
}

type spotifyPartnerTrack struct {
	URI          string                `json:"uri"`
	Name         string                `json:"name"`
	DiscNumber   int                   `json:"discNumber"`
	TrackNumber  int                   `json:"trackNumber"`
	AlbumOfTrack spotifyPartnerAlbum   `json:"albumOfTrack"`
	Artists      spotifyPartnerArtists `json:"artists"`
	Duration     struct {
		TotalMilliseconds int `json:"totalMilliseconds"`
	} `json:"duration"`
	ContentRating struct {
		Label string `json:"label"`
	} `json:"contentRating"`
}

type spotifyPartnerAlbum struct {
	URI      string `json:"uri"`
	Name     string `json:"name"`
	CoverArt struct {
		Sources []SpotifyImage `json:"sources"`
	} `json:"coverArt"`
}

type spotifyPartnerArtists struct {
	Items []struct {
		URI     string `json:"uri"`
		Profile struct {
			Name string `json:"name"`
		} `json:"profile"`
	} `json:"items"`
}

// GetSpotifyPartnerPlaylist fetches Spotify playlist metadata and tracks through Spotify's web partner GraphQL API.
func GetSpotifyPartnerPlaylist(id string) (SpotifyPlaylist, []SpotifyTrack, error) {
	offset := 0
	total := 0
	var playlist SpotifyPlaylist
	tracks := []SpotifyTrack{}

	for {
		page, err := getSpotifyPartnerPlaylistPage(id, spotifyPartnerPageLimit, offset)
		if err != nil {
			return SpotifyPlaylist{}, nil, err
		}
		body := page.Data.PlaylistV2
		if body.ID == "" {
			return SpotifyPlaylist{}, nil, fmt.Errorf("spotify partner playlist %s not found", id)
		}
		if offset == 0 {
			playlist = spotifyPartnerPlaylistToSpotifyPlaylist(page)
			total = body.Content.TotalCount
			playlist.Tracks.Total = total
		}
		for _, item := range body.Content.Items {
			track := spotifyPartnerTrackToSpotifyTrack(item.ItemV2.Data)
			if track.ID == "" || track.Name == "" {
				continue
			}
			tracks = append(tracks, track)
		}
		if body.Content.PagingInfo.NextOffset == nil {
			break
		}
		nextOffset := *body.Content.PagingInfo.NextOffset
		if nextOffset <= offset {
			break
		}
		offset = nextOffset
	}

	return playlist, tracks, nil
}

// GetSpotifyPartnerPlaylistTracks fetches all public playlist tracks through Spotify's web partner GraphQL API.
func GetSpotifyPartnerPlaylistTracks(id string) ([]SpotifyTrack, int, error) {
	playlist, tracks, err := GetSpotifyPartnerPlaylist(id)
	if err != nil {
		return nil, 0, err
	}
	return tracks, playlist.Tracks.Total, nil
}

func getSpotifyPartnerPlaylistPage(id string, limit, offset int) (spotifyPartnerPlaylistResponse, error) {
	var result spotifyPartnerPlaylistResponse

	token, err := getSpotifyAnonymousToken("playlist", id)
	if err != nil {
		return result, err
	}

	variables, err := json.Marshal(map[string]any{
		"uri":    "spotify:playlist:" + id,
		"limit":  limit,
		"offset": offset,
	})
	if err != nil {
		return result, err
	}
	extensions, err := json.Marshal(map[string]any{
		"persistedQuery": map[string]any{
			"version":    1,
			"sha256Hash": spotifyPlaylistQuerySHA,
		},
	})
	if err != nil {
		return result, err
	}

	query := url.Values{}
	query.Set("operationName", spotifyPlaylistQuery)
	query.Set("variables", string(variables))
	query.Set("extensions", string(extensions))

	req, err := http.NewRequest(http.MethodGet, spotifyPartnerQueryURL+"?"+query.Encode(), nil)
	if err != nil {
		return result, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("App-Platform", "WebPlayer")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Spotify-App-Version", "1.2.62.268.gcb6cd226")
	req.Header.Set("User-Agent", spotifyBrowserUserAgent)

	resp, err := spotifyHTTPClient.Do(req)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if resp.StatusCode == http.StatusTooManyRequests {
			if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
				return result, fmt.Errorf("spotify partner API rate limited: retry after %s seconds", retryAfter)
			}
		}
		return result, fmt.Errorf("spotify partner API error: %s", resp.Status)
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return result, err
	}
	return result, nil
}

func spotifyPartnerPlaylistToSpotifyPlaylist(page spotifyPartnerPlaylistResponse) SpotifyPlaylist {
	body := page.Data.PlaylistV2
	playlist := SpotifyPlaylist{
		ID:          body.ID,
		Name:        body.Name,
		Description: body.Description,
		Images:      firstSpotifyImages(body.Images.Items),
		Owner: SpotifyOwner{
			ID:          spotifyIDFromURI(body.OwnerV2.Data.URI),
			DisplayName: body.OwnerV2.Data.Name,
			Type:        "user",
		},
		Type: "playlist",
		URI:  body.URI,
	}
	if playlist.Owner.ID == "" {
		playlist.Owner.ID = body.OwnerV2.Data.Username
	}
	playlist.Tracks.Total = body.Content.TotalCount
	return playlist
}

func spotifyPartnerTrackToSpotifyTrack(track spotifyPartnerTrack) SpotifyTrack {
	artists := make([]SpotifyArtist, 0, len(track.Artists.Items))
	for _, item := range track.Artists.Items {
		if item.Profile.Name == "" {
			continue
		}
		artists = append(artists, SpotifyArtist{
			ID:   spotifyIDFromURI(item.URI),
			Name: item.Profile.Name,
			Type: "artist",
		})
	}

	album := SpotifyAlbumRef{
		ID:     spotifyIDFromURI(track.AlbumOfTrack.URI),
		Name:   track.AlbumOfTrack.Name,
		Images: track.AlbumOfTrack.CoverArt.Sources,
		Type:   "album",
		URI:    track.AlbumOfTrack.URI,
	}
	return SpotifyTrack{
		ID:         spotifyIDFromURI(track.URI),
		Name:       track.Name,
		DurationMS: track.Duration.TotalMilliseconds,
		Explicit:   strings.EqualFold(track.ContentRating.Label, "EXPLICIT"),
		Artists:    artists,
		Album:      album,
		Type:       "track",
		URI:        track.URI,
	}
}

func firstSpotifyImages(items []struct {
	Sources []SpotifyImage `json:"sources"`
}) []SpotifyImage {
	if len(items) == 0 {
		return nil
	}
	return items[0].Sources
}

func spotifyIDFromURI(uri string) string {
	parts := strings.Split(uri, ":")
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}
