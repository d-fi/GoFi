package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/d-fi/GoFi/api"
	"github.com/d-fi/GoFi/download"
	"github.com/d-fi/GoFi/internal/models"
	"github.com/d-fi/GoFi/internal/services/spotify"
	"github.com/d-fi/GoFi/internal/ui"
	internalutils "github.com/d-fi/GoFi/internal/utils"
	"github.com/d-fi/GoFi/types"
	"github.com/d-fi/GoFi/utils"
)

var display = ui.NewDisplayManager()

// downloadHandlerImproved processes downloads with improved UI
func downloadHandlerImproved(url string, downloadPath string, quality int) error {
	ctx := context.Background()
	
	// Parse the URL to identify its type
	parsedInfo, err := internalutils.ParseMusicURL(url)
	if err != nil {
		display.PrintError("Failed to parse URL: %v", err)
		return err
	}

	// Handle Spotify URLs
	if parsedInfo.Source == "spotify" {
		return handleSpotifyDownloadImproved(ctx, parsedInfo, downloadPath, quality)
	}

	// Handle Deezer URLs directly
	if parsedInfo.Source == "deezer" {
		return handleDeezerDownloadImproved(ctx, parsedInfo, downloadPath, quality)
	}

	display.PrintError("Unsupported URL source: %s", parsedInfo.Source)
	return fmt.Errorf("unsupported URL source: %s", parsedInfo.Source)
}

// handleSpotifyDownloadImproved processes Spotify URLs with improved UI
func handleSpotifyDownloadImproved(ctx context.Context, parsedInfo *internalutils.ParsedURLInfo, downloadPath string, quality int) error {
	// Get the Spotify client
	client, _ := getAuthenticatedSpotifyClient(ctx)
	if client == nil {
		display.PrintError("Could not get authenticated Spotify client")
		display.PrintInfo("Run 'gofi auth spotify' to authenticate first")
		return fmt.Errorf("could not get authenticated Spotify client")
	}

	// Create Spotify service
	spotifyService := spotify.NewSpotifyService(client)
	if spotifyService == nil {
		display.PrintError("Failed to initialize Spotify service")
		return fmt.Errorf("failed to initialize Spotify service")
	}

	// Process based on content type
	switch parsedInfo.Type {
	case internalutils.SpotifyTrack:
		return handleSpotifyTrackImproved(ctx, spotifyService, parsedInfo.ID, downloadPath, quality)
	
	case internalutils.SpotifyAlbum:
		return handleSpotifyAlbumImproved(ctx, spotifyService, parsedInfo.ID, downloadPath, quality)
	
	case internalutils.SpotifyPlaylist:
		return handleSpotifyPlaylistImproved(ctx, spotifyService, parsedInfo.ID, downloadPath, quality)
	
	default:
		display.PrintError("Unsupported Spotify content type: %s", parsedInfo.Type)
		return fmt.Errorf("unsupported Spotify content type: %s", parsedInfo.Type)
	}
}

// handleDeezerDownloadImproved processes Deezer URLs with improved UI
func handleDeezerDownloadImproved(ctx context.Context, parsedInfo *internalutils.ParsedURLInfo, downloadPath string, quality int) error {
	// Process based on content type
	switch parsedInfo.Type {
	case internalutils.DeezerTrack:
		return handleDeezerTrackImproved(parsedInfo.ID, downloadPath, quality)
	
	case internalutils.DeezerAlbum:
		return handleDeezerAlbumImproved(parsedInfo.ID, downloadPath, quality)
	
	case internalutils.DeezerPlaylist:
		return handleDeezerPlaylistImproved(parsedInfo.ID, downloadPath, quality)
	
	default:
		display.PrintError("Unsupported Deezer content type: %s", parsedInfo.Type)
		return fmt.Errorf("unsupported Deezer content type: %s", parsedInfo.Type)
	}
}

// handleSpotifyTrackImproved handles downloading a single Spotify track with improved UI
func handleSpotifyTrackImproved(ctx context.Context, spotifyService *spotify.SpotifyService, id string, downloadPath string, quality int) error {
	display.PrintHeader("Spotify Track Download")
	
	// Fetch track from Spotify
	display.PrintSearching("Spotify for track info")
	track, err := spotifyService.FetchTrack(ctx, id)
	if err != nil {
		display.PrintSearchResult(false)
		return fmt.Errorf("failed to fetch track from Spotify: %v", err)
	}
	display.PrintSearchResult(true)

	// Find matching track on Deezer
	display.PrintSearching("Deezer for matching track")
	deezerTrack, err := api.SearchTrackOnDeezer(track)
	if err != nil {
		display.PrintSearchResult(false)
		return fmt.Errorf("failed to find track on Deezer: %v", err)
	}
	display.PrintSearchResult(true)

	// Print track info
	display.PrintTrackInfo(deezerTrack.SNG_TITLE, deezerTrack.ART_NAME, deezerTrack.ALB_TITLE, quality)

	// Create a folder with the artist name
	artistFolder := filepath.Join(downloadPath, deezerTrack.ART_NAME)
	
	// Custom filename for the track: Artist - Title
	customFilename := fmt.Sprintf("%s - %s", deezerTrack.ART_NAME, deezerTrack.SNG_TITLE)

	// Download the track
	return downloadTrackImproved(deezerTrack, artistFolder, quality, customFilename)
}

// handleDeezerTrackImproved handles downloading a single Deezer track with improved UI
func handleDeezerTrackImproved(id string, downloadPath string, quality int) error {
	display.PrintHeader("Deezer Track Download")
	
	display.PrintSearching("Deezer for track info")
	track, err := api.GetTrackInfo(id)
	if err != nil {
		display.PrintSearchResult(false)
		return fmt.Errorf("failed to get track info from Deezer: %v", err)
	}
	display.PrintSearchResult(true)

	// Print track info
	display.PrintTrackInfo(track.SNG_TITLE, track.ART_NAME, track.ALB_TITLE, quality)

	// Create a folder with the artist name
	artistFolder := filepath.Join(downloadPath, track.ART_NAME)
	
	// Custom filename for the track: Artist - Title
	customFilename := fmt.Sprintf("%s - %s", track.ART_NAME, track.SNG_TITLE)

	// Download the track
	return downloadTrackImproved(track, artistFolder, quality, customFilename)
}

// handleSpotifyAlbumImproved handles downloading a Spotify album with improved UI
func handleSpotifyAlbumImproved(ctx context.Context, spotifyService *spotify.SpotifyService, id string, downloadPath string, quality int) error {
	display.PrintHeader("Spotify Album Download")
	
	display.PrintSearching("Spotify for album info")
	album, tracks, err := spotifyService.FetchAlbum(ctx, id)
	if err != nil {
		display.PrintSearchResult(false)
		return fmt.Errorf("failed to fetch album from Spotify: %v", err)
	}
	display.PrintSearchResult(true)

	artistName := joinArtistNames(album.Artists)
	display.PrintAlbumInfo(album.Title, artistName, len(tracks), quality)

	// Find matching album on Deezer
	display.PrintSearching("Deezer for matching album")
	deezerAlbum, err := api.SearchAlbumOnDeezer(album)
	if err != nil {
		display.PrintSearchResult(false)
		display.PrintWarning("Could not find album on Deezer. Trying to match individual tracks...")
		return downloadSpotifyTracksIndividuallyImproved(tracks, downloadPath, quality, "")
	}
	display.PrintSearchResult(true)

	// Create a folder for the album
	albumPath := filepath.Join(downloadPath, deezerAlbum.ALB_TITLE)

	// Get album tracks
	albumTracks, err := api.GetAlbumTracks(fmt.Sprint(deezerAlbum.ALB_ID))
	if err != nil {
		display.PrintError("Failed to get album tracks from Deezer: %v", err)
		return fmt.Errorf("failed to get album tracks from Deezer: %v", err)
	}

	// Download the tracks
	total := len(albumTracks.Data)
	succeeded := 0
	failed := 0

	display.PrintInfo("Starting download of %d tracks...", total)
	fmt.Println()

	for i, track := range albumTracks.Data {
		trackInfo, err := api.GetTrackInfo(fmt.Sprint(track.SNG_ID))
		if err != nil {
			display.PrintError("[%d/%d] Failed to get info for: %s", i+1, total, track.SNG_TITLE)
			failed++
			continue
		}
		
		// Custom filename for the track: Artist - Title
		customFilename := fmt.Sprintf("%s - %s", trackInfo.ART_NAME, trackInfo.SNG_TITLE)
		
		display.PrintInfo("[%d/%d] Downloading: %s", i+1, total, customFilename)
		err = downloadTrackImproved(trackInfo, albumPath, quality, customFilename)
		if err != nil {
			display.PrintError("Failed: %v", err)
			failed++
		} else {
			succeeded++
		}
	}

	display.PrintDownloadSummary(succeeded, failed, total)
	
	if failed > 0 {
		return fmt.Errorf("some tracks failed to download")
	}
	return nil
}

// handleDeezerAlbumImproved handles downloading a Deezer album with improved UI
func handleDeezerAlbumImproved(id string, downloadPath string, quality int) error {
	display.PrintHeader("Deezer Album Download")
	
	display.PrintSearching("Deezer for album info")
	album, err := api.GetAlbumInfo(id)
	if err != nil {
		display.PrintSearchResult(false)
		return fmt.Errorf("failed to get album info from Deezer: %v", err)
	}
	display.PrintSearchResult(true)

	// Get album tracks
	albumTracks, err := api.GetAlbumTracks(id)
	if err != nil {
		display.PrintError("Failed to get album tracks from Deezer: %v", err)
		return fmt.Errorf("failed to get album tracks from Deezer: %v", err)
	}

	display.PrintAlbumInfo(album.ALB_TITLE, album.ART_NAME, len(albumTracks.Data), quality)

	// Create a folder for the album
	albumPath := filepath.Join(downloadPath, album.ALB_TITLE)

	// Download the tracks
	total := len(albumTracks.Data)
	succeeded := 0
	failed := 0

	display.PrintInfo("Starting download of %d tracks...", total)
	fmt.Println()

	for i, track := range albumTracks.Data {
		trackInfo, err := api.GetTrackInfo(fmt.Sprint(track.SNG_ID))
		if err != nil {
			display.PrintError("[%d/%d] Failed to get info for: %s", i+1, total, track.SNG_TITLE)
			failed++
			continue
		}
		
		// Custom filename for the track: Artist - Title
		customFilename := fmt.Sprintf("%s - %s", trackInfo.ART_NAME, trackInfo.SNG_TITLE)
		
		display.PrintInfo("[%d/%d] Downloading: %s", i+1, total, customFilename)
		err = downloadTrackImproved(trackInfo, albumPath, quality, customFilename)
		if err != nil {
			display.PrintError("Failed: %v", err)
			failed++
		} else {
			succeeded++
		}
	}

	display.PrintDownloadSummary(succeeded, failed, total)
	
	if failed > 0 {
		return fmt.Errorf("some tracks failed to download")
	}
	return nil
}

// handleSpotifyPlaylistImproved handles downloading a Spotify playlist with improved UI
func handleSpotifyPlaylistImproved(ctx context.Context, spotifyService *spotify.SpotifyService, id string, downloadPath string, quality int) error {
	display.PrintHeader("Spotify Playlist Download")
	
	display.PrintSearching("Spotify for playlist info")
	playlist, tracks, err := spotifyService.FetchPlaylist(ctx, id)
	if err != nil {
		display.PrintSearchResult(false)
		return fmt.Errorf("failed to fetch playlist from Spotify: %v", err)
	}
	display.PrintSearchResult(true)

	display.PrintPlaylistInfo(playlist.Title, playlist.OwnerName, len(tracks), quality)

	// Create a folder for the playlist using just the playlist name
	playlistPath := filepath.Join(downloadPath, playlist.Title)

	return downloadSpotifyTracksIndividuallyImproved(tracks, playlistPath, quality, "")
}

// handleDeezerPlaylistImproved handles downloading a Deezer playlist with improved UI
func handleDeezerPlaylistImproved(id string, downloadPath string, quality int) error {
	display.PrintHeader("Deezer Playlist Download")
	
	display.PrintSearching("Deezer for playlist info")
	playlist, err := api.GetPlaylistInfo(id)
	if err != nil {
		display.PrintSearchResult(false)
		return fmt.Errorf("failed to get playlist info from Deezer: %v", err)
	}
	display.PrintSearchResult(true)

	// Get playlist tracks
	tracks, err := api.GetPlaylistTracks(id)
	if err != nil {
		display.PrintError("Failed to get playlist tracks from Deezer: %v", err)
		return fmt.Errorf("failed to get playlist tracks from Deezer: %v", err)
	}

	display.PrintPlaylistInfo(playlist.Title, "", len(tracks.Data), quality)

	// Create a folder for the playlist using just the playlist name
	playlistPath := filepath.Join(downloadPath, playlist.Title)

	// Download the tracks
	total := len(tracks.Data)
	succeeded := 0
	failed := 0

	display.PrintInfo("Starting download of %d tracks...", total)
	fmt.Println()

	for i, track := range tracks.Data {
		trackInfo, err := api.GetTrackInfo(fmt.Sprint(track.SNG_ID))
		if err != nil {
			display.PrintError("[%d/%d] Failed to get info for: %s by %s", i+1, total, track.SNG_TITLE, track.ART_NAME)
			failed++
			continue
		}
		
		// Custom filename for the track: Artist - Title
		customFilename := fmt.Sprintf("%s - %s", trackInfo.ART_NAME, trackInfo.SNG_TITLE)
		
		display.PrintInfo("[%d/%d] Downloading: %s", i+1, total, customFilename)
		err = downloadTrackImproved(trackInfo, playlistPath, quality, customFilename)
		if err != nil {
			display.PrintError("Failed: %v", err)
			failed++
		} else {
			succeeded++
		}
	}

	display.PrintDownloadSummary(succeeded, failed, total)

	if failed > 0 {
		return fmt.Errorf("%d out of %d tracks failed to download", failed, total)
	}
	return nil
}

// downloadSpotifyTracksIndividuallyImproved searches for and downloads each track individually with improved UI
func downloadSpotifyTracksIndividuallyImproved(tracks []models.Track, downloadPath string, quality int, playlistName string) error {
	total := len(tracks)
	succeeded := 0
	failed := 0

	display.PrintInfo("Matching %d tracks from Spotify to Deezer...", total)
	fmt.Println()

	for i, track := range tracks {
		// Add a small delay to avoid overwhelming the Deezer API
		if i > 0 && i%3 == 0 {
			time.Sleep(1 * time.Second)
		}

		trackName := fmt.Sprintf("%s by %s", track.Title, joinArtistNames(track.Artists))
		display.PrintInfo("[%d/%d] Searching for: %s", i+1, total, trackName)

		deezerTrack, err := api.SearchTrackOnDeezer(&track)
		if err != nil {
			display.PrintError("Not found on Deezer: %v", err)
			failed++
			continue
		}

		// Always use "Artist - Title" format for all tracks
		customFilename := fmt.Sprintf("%s - %s", deezerTrack.ART_NAME, deezerTrack.SNG_TITLE)

		// Download the track
		err = downloadTrackImproved(deezerTrack, downloadPath, quality, customFilename)
		if err != nil {
			display.PrintError("Download failed: %v", err)
			failed++
			continue
		}

		succeeded++
	}

	display.PrintDownloadSummary(succeeded, failed, total)

	if failed > 0 {
		return fmt.Errorf("%d out of %d tracks failed to download", failed, total)
	}
	return nil
}

// downloadTrackImproved downloads a single track from Deezer with improved UI
func downloadTrackImproved(track types.TrackType, downloadPath string, quality int, customFilename string) error {
	// Determine cover size based on quality
	coverSize := 500
	if quality == 9 {
		coverSize = 1000
	}

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(downloadPath, 0755); err != nil {
		return fmt.Errorf("failed to create download directory: %v", err)
	}

	// Check if file already exists
	ext := "mp3"
	if quality == 9 {
		ext = "flac"
	}
	fullPath := filepath.Join(downloadPath, fmt.Sprintf("%s.%s", utils.SanitizeFileName(customFilename), ext))
	
	if _, err := os.Stat(fullPath); err == nil {
		display.PrintFileExists(filepath.Base(fullPath))
		return nil
	}

	// Create a custom progress callback
	var progressBar *ui.SimpleProgress
	progressCallback := func(progress float64, downloaded, total int64) {
		if progressBar == nil && total > 0 {
			progressBar = display.StartProgress(track.SNG_ID, total, customFilename)
		}
		if progressBar != nil && total > 0 {
			progressBar.Update(downloaded)
		}
	}

	// Create download options
	options := download.DownloadTrackOptions{
		SngID:      fmt.Sprint(track.SNG_ID),
		Quality:    quality,
		CoverSize:  coverSize,
		SaveToDir:  downloadPath,
		Filename:   utils.SanitizeFileName(customFilename),
		OnProgress: progressCallback,
	}

	// Execute download
	_, err := download.DownloadTrack(options)
	if err != nil {
		if progressBar != nil {
			progressBar.Clear()
		}
		return err
	}

	if progressBar != nil {
		display.FinishProgress(track.SNG_ID)
	}

	return nil
}