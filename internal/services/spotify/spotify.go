// Package spotify provides services for interacting with the Spotify API.
package spotify

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http" // Added for status code checking
	"strings"  // Added for joining artists/genres
	"time"

	// Import actual models
	"github.com/d-fi/GoFi/internal/models"
	spotify "github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
)

// --- Placeholder Model Structs ---
// ‼️ FIXME: Replace these with your actual models from internal/models

type SourceType string

const (
	SourceSpotify SourceType = "Spotify"
	SourceDeezer  SourceType = "Deezer"
	// Add other sources as needed
)

type Image struct {
	URL    string
	Height int
	Width  int
}

type Artist struct {
	ID   string // Spotify ID or your internal ID
	Name string
	// Add other fields like Images if needed
}

type Album struct {
	ID          string   // Spotify ID or your internal ID
	Title       string
	Artists     []Artist // List of primary artists
	Images      []Image
	ReleaseDate string   // YYYY-MM-DD or YYYY
	Label       string
	TotalTracks int
	Source      SourceType
	SpotifyID   string `json:",omitempty"` // Ensure this exists if needed elsewhere
	UPC         string `json:",omitempty"` // External ID
	Genres      []string
	AlbumType   string // album, single, compilation
	// Add other relevant fields
}

type Track struct {
	ID          string   // Spotify ID or your internal ID
	Title       string
	Album       string   // Album Title
	AlbumID     string   // Spotify Album ID or your internal ID
	AlbumArtist string   // Primary album artist name
	Artists     []Artist // List of track artists (can differ from album artists)
	DurationMs  int
	Source      SourceType
	SpotifyID   string `json:",omitempty"` // Ensure this exists if needed elsewhere
	ISRC        string `json:",omitempty"` // External ID
	PreviewURL  string `json:",omitempty"`
	Images      []Image  // Typically Album images
	ReleaseDate string   // YYYY-MM-DD or YYYY (from Album)
	TrackNumber int
	DiscNumber  int
	Explicit    bool
	// Add other relevant fields (e.g., DeezerID, Lyrics, FilePath after download)
}

type Playlist struct {
	ID            string // Spotify ID or your internal ID
	Title         string
	Description   string
	OwnerName     string
	OwnerID       string
	Images        []Image
	Source        SourceType
	SpotifyID     string `json:",omitempty"`
	Public        bool
	Collaborative bool
	TotalTracks   int // Approximate, as it might change
	// Add other relevant fields
}

// --- End Placeholder Model Structs ---


// SpotifyService interacts with the Spotify API using an authenticated client.
type SpotifyService struct {
	client *spotify.Client
}

// NewSpotifyService creates a new service instance.
// It requires an authenticated client obtained from AuthService.GetClient().
func NewSpotifyService(client *spotify.Client) *SpotifyService {
	if client == nil {
		// This should ideally not happen if GetClient is used correctly
		log.Println("Warning: Initializing SpotifyService with a nil client.")
		// Return nil or an error? For now, allow it but expect failures later.
		return &SpotifyService{client: nil}
	}
	return &SpotifyService{
		client: client,
	}
}

// --- Mapping Helper Functions ---

func mapSpotifyImageToGofi(img spotify.Image) models.Image {
	return models.Image{
		URL:    img.URL,
		Height: int(img.Height), // Cast Numeric to int
		Width:  int(img.Width),  // Cast Numeric to int
	}
}

func mapSpotifyImagesToGofi(imgs []spotify.Image) []models.Image { // Return []models.Image
	if imgs == nil {
		return nil
	}
	gofiImages := make([]models.Image, len(imgs))
	for i, img := range imgs {
		gofiImages[i] = mapSpotifyImageToGofi(img)
	}
	return gofiImages
}

// Changed to return models.Artist
func mapSpotifyArtistToGofi(artist spotify.SimpleArtist) models.Artist {
	return models.Artist{
		ID:     artist.ID.String(),
		Name:   artist.Name,
		Source: models.SourceSpotify, // Use models constant
		// Images not available on SimpleArtist
	}
}

// Changed to return []models.Artist
func mapSpotifyArtistsToGofi(artists []spotify.SimpleArtist) []models.Artist {
	if artists == nil {
		return nil
	}
	gofiArtists := make([]models.Artist, len(artists)) // Use models.Artist
	for i, artist := range artists {
		gofiArtists[i] = mapSpotifyArtistToGofi(artist)
	}
	return gofiArtists
}

// Maps FullTrack to models.Track
func mapSpotifyFullTrackToGofi(track *spotify.FullTrack) *models.Track {
	if track == nil {
		return nil
	}
	isrc, _ := track.ExternalIDs["isrc"]
	// Map embedded SimpleAlbum to *models.Album
	gofiAlbum := mapSpotifySimpleAlbumToGofi(&track.Album) // Pass the address

	return &models.Track{
		ID:          track.ID.String(),
		SpotifyID:   track.ID.String(), // Populate SpotifyID
		Source:      models.SourceSpotify, // Use models constant
		Title:       track.Name,
		Artists:     mapSpotifyArtistsToGofi(track.Artists), // Returns []models.Artist
		Album:       gofiAlbum,                         // Assign *models.Album
		DurationMs:  int(track.Duration),           // Cast spotify.Numeric to int
		TrackNumber: int(track.TrackNumber),        // Cast spotify.Numeric to int
		DiscNumber:  int(track.DiscNumber),         // Cast spotify.Numeric to int
		Explicit:    track.Explicit,
		ISRC:        isrc,
		Images:      mapSpotifyImagesToGofi(track.Album.Images), // Returns []models.Image
		ReleaseDate: track.Album.ReleaseDate,
		// AddedAt set contextually
	}
}

// Maps SimpleTrack + SimpleAlbum context to models.Track
// Takes *models.Album as context
func mapSpotifySimpleTrackToGofi(track *spotify.SimpleTrack, albumContext *models.Album) *models.Track {
    if track == nil {
        return nil // Return nil if track is nil
    }

    return &models.Track{
        ID:          track.ID.String(),
        SpotifyID:   track.ID.String(), // Populate SpotifyID
        Source:      models.SourceSpotify, // Use models constant
        Title:       track.Name,
        Artists:     mapSpotifyArtistsToGofi(track.Artists), // Returns []models.Artist
        Album:       albumContext,                      // Use provided *models.Album context
        DurationMs:  int(track.Duration),           // Cast spotify.Numeric to int
        TrackNumber: int(track.TrackNumber),        // Cast spotify.Numeric to int
        // DiscNumber not on SimpleTrack
        Explicit:    track.Explicit,
		// ISRC not on SimpleTrack
		// Images taken from albumContext
		// ReleaseDate taken from albumContext
        // AddedAt set contextually
    }
}

// Maps SimpleAlbum to models.Album
func mapSpotifySimpleAlbumToGofi(album *spotify.SimpleAlbum) *models.Album {
	if album == nil {
		return nil
	}
	return &models.Album{
		ID:          album.ID.String(),
		SpotifyID:   album.ID.String(), // Populate SpotifyID
		Source:      models.SourceSpotify,
		Title:       album.Name,
		Artists:     mapSpotifyArtistsToGofi(album.Artists),
		Images:      mapSpotifyImagesToGofi(album.Images),
		ReleaseDate: album.ReleaseDate,
		AlbumType:   album.AlbumType, // Populate AlbumType
		// Label, TotalTracks, UPC, Genres not available on SimpleAlbum
	}
}


// Maps FullAlbum to models.Album
func mapSpotifyAlbumToGofi(album *spotify.FullAlbum) *models.Album {
	if album == nil {
		return nil
	}
	upc, _ := album.ExternalIDs["upc"]
	// Label field is not available in FullAlbum struct
	// Set label to empty string since it doesn't exist in the Spotify API response
	label := "" // Label is not available in the FullAlbum struct
	// Cast Total via embedded Tracks field
	totalTracks := int(album.Tracks.Total) // Cast spotify.Numeric to int


	return &models.Album{
		ID:          album.ID.String(),
		SpotifyID:   album.ID.String(),   // Populate SpotifyID
		Source:      models.SourceSpotify, // Use models constant
		Title:       album.Name,
		Artists:     mapSpotifyArtistsToGofi(album.Artists), // Returns []models.Artist
		Images:      mapSpotifyImagesToGofi(album.Images), // Returns []models.Image
		ReleaseDate: album.ReleaseDate,
		Label:       label,             // Populate Label (direct access)
		TotalTracks: totalTracks,       // Populate TotalTracks (casted)
		UPC:         upc,
		Genres:      album.Genres,      // Populate Genres
		AlbumType:   album.AlbumType,   // Populate AlbumType
	}
}

// Maps FullPlaylist to models.Playlist
func mapSpotifyPlaylistToGofi(playlist *spotify.FullPlaylist) *models.Playlist {
	if playlist == nil {
		return nil
	}
    // Access IsPublic (not Public) field from the embedded SimplePlaylist
    isPublic := playlist.IsPublic // Correct field name in the library
	// Cast Total via embedded Tracks field
	totalTracks := int(playlist.Tracks.Total) // Cast spotify.Numeric to int
	// Access DisplayName via embedded Owner field
	ownerName := playlist.Owner.DisplayName


	return &models.Playlist{
		ID:          playlist.ID.String(),
		SpotifyID:   playlist.ID.String(),   // Populate SpotifyID
		Source:      models.SourceSpotify, // Use models constant
		Title:       playlist.Name,
		Description: playlist.Description,
		OwnerName:   ownerName, // Populate OwnerName (direct access)
		Images:      mapSpotifyImagesToGofi(playlist.Images), // Returns []models.Image
		Public:      isPublic,          // Populate Public (direct access)
		TotalTracks: totalTracks,       // Populate TotalTracks (casted)
		// Removed OwnerID, Collaborative as they aren't in models.Playlist
	}
}


// --- Service Methods ---

// FetchTrack retrieves a single track's details from Spotify.
func (s *SpotifyService) FetchTrack(ctx context.Context, id string) (*models.Track, error) { // Return *models.Track
	if s.client == nil {
		return nil, fmt.Errorf("spotify client not initialized")
	}
	trackID := spotify.ID(id)
	fullTrack, err := s.client.GetTrack(ctx, trackID)
	if err != nil {
		if spotifyErr, ok := err.(*spotify.Error); ok && spotifyErr.Status == http.StatusNotFound {
			return nil, fmt.Errorf("spotify track %s not found", id)
		}
		return nil, fmt.Errorf("failed to get spotify track %s: %w", id, err)
	}

	gofiTrack := mapSpotifyFullTrackToGofi(fullTrack) // Returns *models.Track
	if gofiTrack == nil {
		 return nil, fmt.Errorf("failed to map spotify track %s", id)
	}

	log.Printf("Fetched Spotify track: %s - %s", gofiTrack.Title, joinArtists(gofiTrack.Artists)) // Use models.Artist slice
	return gofiTrack, nil
}


// FetchAlbum retrieves album details and its tracks from Spotify.
func (s *SpotifyService) FetchAlbum(ctx context.Context, id string) (*models.Album, []models.Track, error) {
	if s.client == nil {
		return nil, nil, fmt.Errorf("spotify client not initialized")
	}
	albumID := spotify.ID(id)

	fullAlbum, err := s.client.GetAlbum(ctx, albumID)
	if err != nil {
		if spotifyErr, ok := err.(*spotify.Error); ok && spotifyErr.Status == http.StatusNotFound {
			return nil, nil, fmt.Errorf("spotify album %s not found", id)
		}
		return nil, nil, fmt.Errorf("failed to get spotify album %s: %w", id, err)
	}

	gofiAlbum := mapSpotifyAlbumToGofi(fullAlbum) // Returns *models.Album
    if gofiAlbum == nil {
        return nil, nil, fmt.Errorf("failed to map spotify album %s", id)
    }

	log.Printf("Fetching tracks for Spotify album: %s (%s)", gofiAlbum.Title, id)
	var gofiTracks []models.Track // Use models.Track slice
	offset := 0
	limit := 50
	// Map the SimpleAlbum part for context, now returns *models.Album
	simpleAlbumForMapping := mapSpotifySimpleAlbumToGofi(&fullAlbum.SimpleAlbum)

	for {
		albumTracksPage, err := s.client.GetAlbumTracks(ctx, albumID, spotify.Offset(offset), spotify.Limit(limit))
		if err != nil {
			log.Printf("Error fetching album tracks page (offset %d) for %s: %v. Returning partially fetched data.", offset, id, err)
			return gofiAlbum, gofiTracks, fmt.Errorf("failed to fetch all album tracks for %s (page offset %d): %w", id, offset, err)
		}
		if albumTracksPage == nil || len(albumTracksPage.Tracks) == 0 {
			break
		}

		log.Printf("Fetched page of %d tracks for album %s (offset %d)", len(albumTracksPage.Tracks), id, offset)
		for _, simpleTrack := range albumTracksPage.Tracks {
			// Pass *models.Album context
			gofiTrack := mapSpotifySimpleTrackToGofi(&simpleTrack, simpleAlbumForMapping) // Returns *models.Track
            if gofiTrack != nil {
			    gofiTracks = append(gofiTracks, *gofiTrack) // Append models.Track
            } else {
                log.Printf("Warning: Failed to map simple track %s from album %s", simpleTrack.ID, id)
            }
		}

		if len(albumTracksPage.Tracks) < limit {
			break
		}
		offset += len(albumTracksPage.Tracks)
	}

    log.Printf("Fetched total %d tracks for Spotify album: %s - %s", len(gofiTracks), gofiAlbum.Title, joinArtists(gofiAlbum.Artists)) // Use models.Artist slice
	return gofiAlbum, gofiTracks, nil
}


// FetchPlaylist retrieves playlist details and its tracks from Spotify.
func (s *SpotifyService) FetchPlaylist(ctx context.Context, id string) (*models.Playlist, []models.Track, error) {
	if s.client == nil {
		return nil, nil, fmt.Errorf("spotify client not initialized")
	}
	playlistID := spotify.ID(id)

	fullPlaylist, err := s.client.GetPlaylist(ctx, playlistID)
	if err != nil {
		if spotifyErr, ok := err.(*spotify.Error); ok && spotifyErr.Status == http.StatusNotFound {
			return nil, nil, fmt.Errorf("spotify playlist %s not found", id)
		}
		return nil, nil, fmt.Errorf("failed to get spotify playlist %s: %w", id, err)
	}

	gofiPlaylist := mapSpotifyPlaylistToGofi(fullPlaylist) // Returns *models.Playlist
    if gofiPlaylist == nil {
         return nil, nil, fmt.Errorf("failed to map spotify playlist %s", id)
    }

	log.Printf("Fetching items for Spotify playlist: %s (%s)", gofiPlaylist.Title, id)
	var gofiTracks []models.Track // Use models.Track slice
	offset := 0
	limit := 100

	for {
		playlistItemsPage, err := s.client.GetPlaylistItems(ctx, playlistID, spotify.Offset(offset), spotify.Limit(limit))
		if err != nil {
			log.Printf("Error fetching playlist items page (offset %d) for %s: %v. Returning partially fetched data.", offset, id, err)
			return gofiPlaylist, gofiTracks, fmt.Errorf("failed to fetch all playlist items for %s (page offset %d): %w", id, offset, err)
		}
		if playlistItemsPage == nil || len(playlistItemsPage.Items) == 0 {
			break
		}

		log.Printf("Fetched page of %d items for playlist %s (offset %d)", len(playlistItemsPage.Items), id, offset)
		for _, item := range playlistItemsPage.Items {
			if item.Track.Track != nil {
				gofiTrack := mapSpotifyFullTrackToGofi(item.Track.Track) // Returns *models.Track
                if gofiTrack != nil {
					// Check if AddedAt is not empty before parsing
					if item.AddedAt != "" {
						// Parse the timestamp string directly using time.Parse with RFC3339 format
						parsedTime, err := time.Parse(time.RFC3339, item.AddedAt)
						if err != nil {
							log.Printf("Warning: Failed to parse AddedAt timestamp '%s' for track %s in playlist %s: %v", item.AddedAt, gofiTrack.ID, id, err)
						} else {
							gofiTrack.AddedAt = &parsedTime
						}
					}
				    gofiTracks = append(gofiTracks, *gofiTrack) // Append models.Track
                } else {
                     log.Printf("Warning: Failed to map track %s from playlist %s", item.Track.Track.ID, id)
                }
			} else {
                itemType := "unknown"
                if item.Track.Episode != nil { itemType = "episode" }
				log.Printf("Skipping non-track item (type: %s) in playlist %s", itemType, id)
			}
		}

		if len(playlistItemsPage.Items) < limit {
			break
		}
		offset += len(playlistItemsPage.Items)
	}

    log.Printf("Fetched total %d tracks for Spotify playlist: %s", len(gofiTracks), gofiPlaylist.Title)
	return gofiPlaylist, gofiTracks, nil
}

// Simple helper to join artist names for logging (now takes models.Artist)
func joinArtists(artists []models.Artist) string {
    if len(artists) == 0 {
		return "Unknown Artist(s)"
	}
    names := make([]string, 0, len(artists))
    for _, a := range artists {
		if a.Name != "" {
			names = append(names, a.Name)
		}
    }
	if len(names) == 0 {
        return "Unknown Artist(s)" // Handle case where all names were empty
    }
    return strings.Join(names, ", ")
}

// GetClientFromToken retrieves a Spotify client using a stored token.
// Assumes Config struct (defined elsewhere, e.g., auth.go) has OAuthConfig() method.
func GetClientFromToken(ctx context.Context, token *oauth2.Token, config *Config) (*spotify.Client, error) {
	if config == nil {
		return nil, errors.New("spotify config cannot be nil")
	}
	
	oauthCfg := config.OAuthConfig()
	if oauthCfg == nil {
		return nil, errors.New("oauth2 config within spotify config is nil")
	}
	httpClient := oauthCfg.Client(ctx, token)
	client := spotify.New(httpClient)
	return client, nil
}
