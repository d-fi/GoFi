// cmd/gofi/cmd/download.go (Placeholder Example)
package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/d-fi/GoFi/internal/models"
	"github.com/d-fi/GoFi/internal/services/spotify"
	"github.com/d-fi/GoFi/internal/utils"
	"github.com/spf13/cobra"
	spotifyClient "github.com/zmb3/spotify/v2" // Alias import to avoid naming conflict
)

// downloadCmd represents the download command (placeholder)
var downloadCmd = &cobra.Command{
	Use:   "download [url]",
	Short: "Download music from a given URL (Spotify supported)",
	Long:  `Downloads tracks, albums, or playlists from supported services like Spotify.`,
	Args:  cobra.ExactArgs(1), // Requires exactly one argument: the URL
	Run: func(cmd *cobra.Command, args []string) {
		url := args[0]
		fmt.Printf("Attempting to process URL: %s\n", url)
		handleDownload(url) // Call the handler logic
	},
}

func init() {
	// ‼️ FIXME: This registration assumes 'rootCmd' exists. Adapt as needed.
	// Register this placeholder command with your actual root command.
	// For example: rootCmd.AddCommand(downloadCmd)
	log.Println("Placeholder download command initialized (needs registration with rootCmd)")
}

// --- Integration Logic ---

func handleDownload(url string) {
	ctx := context.Background()

	parsedInfo, err := utils.ParseMusicURL(url)
	if err != nil {
		log.Fatalf("Error: Failed to parse URL: %v", err)
		return
	}

	// Only proceed if we have a valid Spotify client
	client, _ := getAuthenticatedSpotifyClient(ctx) // Error handled within
	if client == nil {
		// Check if the URL was actually a Spotify URL before failing
		if parsedInfo.Source == "spotify" {
			log.Fatal("Error: Failed to get authenticated Spotify client. Cannot proceed with Spotify download. Hint: Run 'gofi auth spotify' first.")
		} else {
			// Log non-fatal error if it wasn't a spotify URL to begin with
			log.Printf("Info: Spotify client not available, but URL source is '%s'. Skipping Spotify check.", parsedInfo.Source)
		}
		// If not spotify, allow falling through to default or other handlers
		// If it *was* spotify, we exit here.
		if parsedInfo.Source == "spotify" {
			return
		}
	}

	// Initialize spotifyService only if client is not nil (meaning auth was successful or not needed)
	var spotifyService *spotify.SpotifyService
	if client != nil {
		spotifyService = spotify.NewSpotifyService(client)
	}

	switch parsedInfo.Type {
	case utils.SpotifyTrack:
		if spotifyService == nil { // Defensive check
			log.Fatal("Error: Spotify service not initialized. Cannot fetch track.")
			return
		}
		log.Printf("Processing Spotify Track ID: %s", parsedInfo.ID)
		track, err := spotifyService.FetchTrack(ctx, parsedInfo.ID)
		if err != nil {
			log.Fatalf("Error fetching Spotify track %s: %v", parsedInfo.ID, err)
			return
		}
		// ‼️ FIXME: Replace placeholder log with actual download logic for 'track'
		log.Printf("=== Fetched Track (Ready for Download) ===")
		log.Printf("  Title: %s", track.Title)
		log.Printf("  Artist(s): %s", joinModelArtists(track.Artists))
		// Ensure track.Album is not nil before accessing Title
		albumTitle := "Unknown Album"
		if track.Album != nil {
			albumTitle = track.Album.Title
		}
		log.Printf("  Album: %s", albumTitle)
		log.Printf("  ISRC: %s", track.ISRC)
		log.Println("------------------------------------------")
		// downloadSingleTrack(track) // Your download function here

	case utils.SpotifyAlbum:
		if spotifyService == nil { // Defensive check
			log.Fatal("Error: Spotify service not initialized. Cannot fetch album.")
			return
		}
		log.Printf("Processing Spotify Album ID: %s", parsedInfo.ID)
		album, tracks, err := spotifyService.FetchAlbum(ctx, parsedInfo.ID)
		if err != nil {
			log.Fatalf("Error fetching Spotify album %s: %v", parsedInfo.ID, err)
			return
		}
		log.Printf("=== Fetched Album (Ready for Download) ===")
		log.Printf("  Title: %s", album.Title)
		log.Printf("  Artist(s): %s", joinModelArtists(album.Artists))
		log.Printf("  Total Tracks: %d (fetched %d)", album.TotalTracks, len(tracks))
		log.Println("----------------------------------------")
		// ‼️ FIXME: Replace placeholder log with actual download logic for all 'tracks'
		log.Printf("Processing %d tracks for album download:", len(tracks))
		for i, track := range tracks {
			log.Printf("  %d. %s - %s", i+1, track.Title, joinModelArtists(track.Artists))
			// downloadSingleTrack(track) // Call download for each track
		}

	case utils.SpotifyPlaylist:
		if spotifyService == nil { // Defensive check
			log.Fatal("Error: Spotify service not initialized. Cannot fetch playlist.")
			return
		}
		log.Printf("Processing Spotify Playlist ID: %s", parsedInfo.ID)
		playlist, tracks, err := spotifyService.FetchPlaylist(ctx, parsedInfo.ID)
		if err != nil {
			log.Fatalf("Error fetching Spotify playlist %s: %v", parsedInfo.ID, err)
			return
		}
		log.Printf("=== Fetched Playlist (Ready for Download) ===")
		log.Printf("  Title: %s", playlist.Title)
		log.Printf("  Owner: %s", playlist.OwnerName)
		log.Printf("  Total Tracks: %d (fetched %d)", playlist.TotalTracks, len(tracks))
		log.Println("-------------------------------------------")
		// ‼️ FIXME: Replace placeholder log with actual download logic for all 'tracks'
		log.Printf("Processing %d tracks for playlist download:", len(tracks))
		for i, track := range tracks {
			log.Printf("  %d. %s - %s", i+1, track.Title, joinModelArtists(track.Artists))
			// downloadSingleTrack(track) // Call download for each track
		}

	// case utils.DeezerTrack: // Example for future expansion
	// 	log.Printf("Processing Deezer Track ID: %s", parsedInfo.ID)
	// 	// ... handle Deezer ...

	default:
		// Check if source was identified before saying unsupported
		if parsedInfo.Source != "" {
			log.Printf("Error: URL type '%s' from service '%s' is not supported for download yet.", parsedInfo.Type, parsedInfo.Source)
		} else {
			log.Printf("Error: Could not determine music service or type from the provided URL.")
		}
	}
}

// Helper function to get authenticated client (handles prompting for auth)
// Returns nil client if auth fails or is not configured.
func getAuthenticatedSpotifyClient(ctx context.Context) (*spotifyClient.Client, *spotify.AuthService) {
	clientID := os.Getenv("SPOTIFY_CLIENT_ID")
	clientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		log.Println("Info: SPOTIFY_CLIENT_ID and SPOTIFY_CLIENT_SECRET environment variables not set. Spotify features disabled.")
		return nil, nil
	}
	cfg := spotify.Config{ClientID: clientID, ClientSecret: clientSecret}
	authService, err := spotify.NewAuthService(cfg)
	if err != nil {
		log.Printf("Error creating Spotify auth service: %v", err)
		return nil, nil
	}

	client, err := authService.GetClient(ctx) // Attempts to load token or refresh
	if err != nil {
		// This is not necessarily an error, just means we need to authenticate.
		// Logged as Info level. The calling function decides if it's fatal.
		log.Printf("Info: Could not automatically retrieve Spotify token (may need initial auth): %v", err)
		return nil, authService // Return authService even if client is nil
	}
	log.Println("Successfully obtained authenticated Spotify client.")
	return client, authService
}

// Helper to join artist names from models.Artist slice
func joinModelArtists(artists []models.Artist) string { // Use the actual model type
	if len(artists) == 0 {
		return "Unknown Artist"
	}
	names := make([]string, len(artists))
	for i, a := range artists {
		if a.Name != "" { // Avoid adding empty names
			names[i] = a.Name
		} else {
			names[i] = "Unnamed Artist" // Placeholder if name is empty
		}
	}
	// Filter out empty strings before joining, in case the above logic missed something
	filteredNames := []string{}
	for _, name := range names {
		if name != "" && name != "Unnamed Artist" { // filter potentially empty strings
			filteredNames = append(filteredNames, name)
		}
	}
	if len(filteredNames) == 0 { // If all artists were unnamed/empty
		return "Unknown Artist(s)"
	}
	return strings.Join(filteredNames, ", ")
} 