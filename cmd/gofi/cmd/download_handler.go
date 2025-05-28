package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/d-fi/GoFi/api"
	"github.com/d-fi/GoFi/download"
	"github.com/d-fi/GoFi/internal/models"
	"github.com/d-fi/GoFi/internal/services/spotify"
	internalutils "github.com/d-fi/GoFi/internal/utils"
	"github.com/d-fi/GoFi/types"
	"github.com/d-fi/GoFi/utils"
)

// downloadHandler processes downloads based on URL type
func downloadHandler(url string, downloadPath string, quality int) error {
	ctx := context.Background()
	
	// Parse the URL to identify its type
	parsedInfo, err := internalutils.ParseMusicURL(url)
	if err != nil {
		return fmt.Errorf("failed to parse URL: %v", err)
	}

	// Handle Spotify URLs
	if parsedInfo.Source == "spotify" {
		return handleSpotifyDownload(ctx, parsedInfo, downloadPath, quality)
	}

	// Handle Deezer URLs directly
	if parsedInfo.Source == "deezer" {
		return handleDeezerDownload(ctx, parsedInfo, downloadPath, quality)
	}

	return fmt.Errorf("unsupported URL source: %s", parsedInfo.Source)
}

// handleSpotifyDownload processes Spotify URLs and downloads content from Deezer
func handleSpotifyDownload(ctx context.Context, parsedInfo *internalutils.ParsedURLInfo, downloadPath string, quality int) error {
	// Get the Spotify client
	client, _ := getAuthenticatedSpotifyClient(ctx)
	if client == nil {
		return fmt.Errorf("could not get authenticated Spotify client - run 'gofi auth spotify' first")
	}

	// Create Spotify service
	spotifyService := spotify.NewSpotifyService(client)
	if spotifyService == nil {
		return fmt.Errorf("failed to initialize Spotify service")
	}

	// Process based on content type
	switch parsedInfo.Type {
	case internalutils.SpotifyTrack:
		return handleSpotifyTrack(ctx, spotifyService, parsedInfo.ID, downloadPath, quality)
	
	case internalutils.SpotifyAlbum:
		return handleSpotifyAlbum(ctx, spotifyService, parsedInfo.ID, downloadPath, quality)
	
	case internalutils.SpotifyPlaylist:
		return handleSpotifyPlaylist(ctx, spotifyService, parsedInfo.ID, downloadPath, quality)
	
	default:
		return fmt.Errorf("unsupported Spotify content type: %s", parsedInfo.Type)
	}
}

// handleDeezerDownload processes Deezer URLs and downloads content directly
func handleDeezerDownload(ctx context.Context, parsedInfo *internalutils.ParsedURLInfo, downloadPath string, quality int) error {
	// Process based on content type
	switch parsedInfo.Type {
	case internalutils.DeezerTrack:
		return handleDeezerTrack(parsedInfo.ID, downloadPath, quality)
	
	case internalutils.DeezerAlbum:
		return handleDeezerAlbum(parsedInfo.ID, downloadPath, quality)
	
	case internalutils.DeezerPlaylist:
		return handleDeezerPlaylist(parsedInfo.ID, downloadPath, quality)
	
	default:
		return fmt.Errorf("unsupported Deezer content type: %s", parsedInfo.Type)
	}
}

// handleDeezerTrack handles downloading a single Deezer track
func handleDeezerTrack(id string, downloadPath string, quality int) error {
	fmt.Printf("Getting track info from Deezer... ")
	track, err := api.GetTrackInfo(id)
	if err != nil {
		return fmt.Errorf("failed to get track info from Deezer: %v", err)
	}
	fmt.Printf("✓\n")

	// Print track info
	fmt.Printf("\nTrack info:\n")
	fmt.Printf("  Title: %s\n", track.SNG_TITLE)
	fmt.Printf("  Artist: %s\n", track.ART_NAME)
	fmt.Printf("  Album: %s\n", track.ALB_TITLE)
	fmt.Printf("  Quality: %d\n", quality)
	fmt.Printf("\nDownloading track...\n")

	// Create a folder with the artist name
	artistFolder := filepath.Join(downloadPath, track.ART_NAME)
	
	// Custom filename for the track: Artist - Title
	customFilename := fmt.Sprintf("%s - %s", track.ART_NAME, track.SNG_TITLE)

	// Download the track
	return downloadTrack(track, artistFolder, quality, customFilename)
}

// handleDeezerAlbum handles downloading a Deezer album
func handleDeezerAlbum(id string, downloadPath string, quality int) error {
	fmt.Printf("Getting album info from Deezer... ")
	album, err := api.GetAlbumInfo(id)
	if err != nil {
		return fmt.Errorf("failed to get album info from Deezer: %v", err)
	}
	fmt.Printf("✓\n")

	// Print album info
	fmt.Printf("\nAlbum info:\n")
	fmt.Printf("  Title: %s\n", album.ALB_TITLE)
	fmt.Printf("  Artist: %s\n", album.ART_NAME)
	fmt.Printf("  Quality: %d\n", quality)
	fmt.Printf("\nDownloading album...\n")

	// Create a folder for the album
	albumPath := filepath.Join(downloadPath, album.ALB_TITLE)

	// Get album tracks
	albumTracks, err := api.GetAlbumTracks(id)
	if err != nil {
		return fmt.Errorf("failed to get album tracks from Deezer: %v", err)
	}

	// Download the tracks
	var lastError error
	for _, track := range albumTracks.Data {
		trackInfo, err := api.GetTrackInfo(fmt.Sprint(track.SNG_ID))
		if err != nil {
			fmt.Printf("Error getting info for track %s: %v\n", track.SNG_TITLE, err)
			lastError = err
			continue
		}
		
		// Custom filename for the track: Artist - Title
		customFilename := fmt.Sprintf("%s - %s", trackInfo.ART_NAME, trackInfo.SNG_TITLE)
		
		err = downloadTrack(trackInfo, albumPath, quality, customFilename)
		if err != nil {
			fmt.Printf("Error downloading track %s: %v\n", track.SNG_TITLE, err)
			lastError = err
		}
	}

	if lastError != nil {
		return fmt.Errorf("some tracks failed to download")
	}
	return nil
}

// handleDeezerPlaylist handles downloading a Deezer playlist
func handleDeezerPlaylist(id string, downloadPath string, quality int) error {
	fmt.Printf("Getting playlist info from Deezer... ")
	playlist, err := api.GetPlaylistInfo(id)
	if err != nil {
		return fmt.Errorf("failed to get playlist info from Deezer: %v", err)
	}
	fmt.Printf("✓\n")

	// Get playlist tracks
	tracks, err := api.GetPlaylistTracks(id)
	if err != nil {
		return fmt.Errorf("failed to get playlist tracks from Deezer: %v", err)
	}

	fmt.Printf("Found playlist: %s (%d tracks)\n", playlist.Title, len(tracks.Data))

	// Create a folder for the playlist using just the playlist name
	playlistPath := filepath.Join(downloadPath, playlist.Title)

	// Download the tracks
	total := len(tracks.Data)
	succeeded := 0
	failed := 0

	for i, track := range tracks.Data {
		fmt.Printf("[%d/%d] Processing: %s by %s... ", 
			i+1, total, track.SNG_TITLE, track.ART_NAME)

		trackInfo, err := api.GetTrackInfo(fmt.Sprint(track.SNG_ID))
		if err != nil {
			fmt.Printf("✗\n")
			fmt.Printf("     Failed to get track info: %v\n", err)
			failed++
			continue
		}
		
		// Custom filename for the track: Artist - Title
		customFilename := fmt.Sprintf("%s - %s", trackInfo.ART_NAME, trackInfo.SNG_TITLE)
		
		err = downloadTrack(trackInfo, playlistPath, quality, customFilename)
		if err != nil {
			fmt.Printf("✗\n")
			fmt.Printf("     Error downloading: %v\n", err)
			failed++
			continue
		}

		succeeded++
	}

	fmt.Printf("\nDownload summary: %d succeeded, %d failed out of %d total\n", 
		succeeded, failed, total)

	if failed > 0 {
		return fmt.Errorf("%d out of %d tracks failed to download", failed, total)
	}
	return nil
}

// handleSpotifyTrack handles downloading a single Spotify track
func handleSpotifyTrack(ctx context.Context, spotifyService *spotify.SpotifyService, id string, downloadPath string, quality int) error {
	fmt.Printf("Getting track info from Spotify... ")
	track, err := spotifyService.FetchTrack(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to fetch track from Spotify: %v", err)
	}
	fmt.Printf("✓\n")

	// Find matching track on Deezer
	fmt.Printf("Searching for track on Deezer... ")
	deezerTrack, err := api.SearchTrackOnDeezer(track)
	if err != nil {
		return fmt.Errorf("failed to find track on Deezer: %v", err)
	}
	fmt.Printf("✓\n")

	// Print found track info
	fmt.Printf("\nFound track on Deezer:\n")
	fmt.Printf("  Title: %s\n", deezerTrack.SNG_TITLE)
	fmt.Printf("  Artist: %s\n", deezerTrack.ART_NAME)
	fmt.Printf("  Album: %s\n", deezerTrack.ALB_TITLE)
	fmt.Printf("  Quality: %d\n", quality)
	fmt.Printf("\nDownloading track...\n")

	// Create a folder with the artist name
	artistFolder := filepath.Join(downloadPath, deezerTrack.ART_NAME)
	
	// Custom filename for the track: Artist - Title
	customFilename := fmt.Sprintf("%s - %s", deezerTrack.ART_NAME, deezerTrack.SNG_TITLE)

	// Download the track
	return downloadTrack(deezerTrack, artistFolder, quality, customFilename)
}

// handleSpotifyAlbum handles downloading a Spotify album
func handleSpotifyAlbum(ctx context.Context, spotifyService *spotify.SpotifyService, id string, downloadPath string, quality int) error {
	fmt.Printf("Getting album info from Spotify... ")
	album, tracks, err := spotifyService.FetchAlbum(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to fetch album from Spotify: %v", err)
	}
	fmt.Printf("✓\n")

	fmt.Printf("Found album: %s by %s (%d tracks)\n", 
		album.Title, 
		joinArtistNames(album.Artists), 
		len(tracks))

	// Find matching album on Deezer
	fmt.Printf("Searching for album on Deezer... ")
	deezerAlbum, err := api.SearchAlbumOnDeezer(album)
	if err != nil {
		fmt.Printf("✗\n")
		fmt.Printf("Could not find album on Deezer. Trying to match individual tracks...\n")
		return downloadSpotifyTracksIndividually(tracks, downloadPath, quality, "")
	}
	fmt.Printf("✓\n")

	// Print found album info
	fmt.Printf("\nFound album on Deezer:\n")
	fmt.Printf("  Title: %s\n", deezerAlbum.ALB_TITLE)
	fmt.Printf("  Artist: %s\n", deezerAlbum.ART_NAME)
	fmt.Printf("  Quality: %d\n", quality)
	fmt.Printf("\nDownloading album...\n")

	// Create a folder for the album
	albumPath := filepath.Join(downloadPath, deezerAlbum.ALB_TITLE)

	// Get album tracks
	albumTracks, err := api.GetAlbumTracks(fmt.Sprint(deezerAlbum.ALB_ID))
	if err != nil {
		return fmt.Errorf("failed to get album tracks from Deezer: %v", err)
	}

	// Download the tracks
	var lastError error
	for _, track := range albumTracks.Data {
		trackInfo, err := api.GetTrackInfo(fmt.Sprint(track.SNG_ID))
		if err != nil {
			fmt.Printf("Error getting info for track %s: %v\n", track.SNG_TITLE, err)
			lastError = err
			continue
		}
		
		// Custom filename for the track: Artist - Title
		customFilename := fmt.Sprintf("%s - %s", trackInfo.ART_NAME, trackInfo.SNG_TITLE)
		
		err = downloadTrack(trackInfo, albumPath, quality, customFilename)
		if err != nil {
			fmt.Printf("Error downloading track %s: %v\n", track.SNG_TITLE, err)
			lastError = err
		}
	}

	if lastError != nil {
		return fmt.Errorf("some tracks failed to download")
	}
	return nil
}

// handleSpotifyPlaylist handles downloading a Spotify playlist
func handleSpotifyPlaylist(ctx context.Context, spotifyService *spotify.SpotifyService, id string, downloadPath string, quality int) error {
	fmt.Printf("Getting playlist info from Spotify... ")
	playlist, tracks, err := spotifyService.FetchPlaylist(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to fetch playlist from Spotify: %v", err)
	}
	fmt.Printf("✓\n")

	fmt.Printf("Found playlist: %s by %s (%d tracks)\n", 
		playlist.Title, 
		playlist.OwnerName, 
		len(tracks))

	// Create a folder for the playlist using just the playlist name
	playlistPath := filepath.Join(downloadPath, playlist.Title)

	return downloadSpotifyTracksIndividually(tracks, playlistPath, quality, "")
}

// downloadSpotifyTracksIndividually searches for and downloads each track individually
func downloadSpotifyTracksIndividually(tracks []models.Track, downloadPath string, quality int, playlistName string) error {
	total := len(tracks)
	succeeded := 0
	failed := 0

	fmt.Printf("Matching %d tracks from Spotify to Deezer...\n", total)

	for i, track := range tracks {
		// Add a small delay to avoid overwhelming the Deezer API
		if i > 0 && i%3 == 0 {
			time.Sleep(1 * time.Second)
		}

		fmt.Printf("[%d/%d] Processing: %s by %s... ", 
			i+1, total, track.Title, joinArtistNames(track.Artists))

		deezerTrack, err := api.SearchTrackOnDeezer(&track)
		if err != nil {
			fmt.Printf("✗\n")
			fmt.Printf("     Failed to find on Deezer: %v\n", err)
			failed++
			continue
		}
		fmt.Printf("✓\n")

		// Always use "Artist - Title" format for all tracks
		customFilename := fmt.Sprintf("%s - %s", deezerTrack.ART_NAME, deezerTrack.SNG_TITLE)

		// Download the track
		err = downloadTrack(deezerTrack, downloadPath, quality, customFilename)
		if err != nil {
			fmt.Printf("     Error downloading: %v\n", err)
			failed++
			continue
		}

		succeeded++
	}

	fmt.Printf("\nDownload summary: %d succeeded, %d failed out of %d total\n", 
		succeeded, failed, total)

	if failed > 0 {
		return fmt.Errorf("%d out of %d tracks failed to download", failed, total)
	}
	return nil
}

// downloadTrack downloads a single track from Deezer
func downloadTrack(track types.TrackType, downloadPath string, quality int, customFilename string) error {
	// Determine cover size based on quality
	// For FLAC (quality=9), use 1000
	// For MP3 320 (quality=3), use 500
	// For MP3 128 (quality=1), use 500
	coverSize := 500
	if quality == 9 {
		coverSize = 1000
	}

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(downloadPath, 0755); err != nil {
		return fmt.Errorf("failed to create download directory: %v", err)
	}

	// Create download options
	options := download.DownloadTrackOptions{
		SngID:     fmt.Sprint(track.SNG_ID),
		Quality:   quality,
		CoverSize: coverSize, // Use the appropriate cover size based on quality
		SaveToDir: downloadPath,
		Filename:  utils.SanitizeFileName(customFilename), // Use a custom filename without the ID suffix
		OnProgress: func(progress float64, _, _ int64) {
			// Simple progress indicator
			if int(progress)%10 == 0 {
				fmt.Printf(".")
			}
		},
	}

	// Execute download
	filePath, err := download.DownloadTrack(options)
	if err != nil {
		return err
	}

	fmt.Printf(" Saved to: %s\n", filePath)
	return nil
}

// joinArtistNames combines artist names into a comma-separated string
func joinArtistNames(artists []models.Artist) string {
	if len(artists) == 0 {
		return "Unknown Artist"
	}

	names := make([]string, 0, len(artists))
	for _, artist := range artists {
		if artist.Name != "" {
			names = append(names, artist.Name)
		}
	}

	if len(names) == 0 {
		return "Unknown Artist"
	}

	return strings.Join(names, ", ")
}