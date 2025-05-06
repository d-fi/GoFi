package api

import (
	"fmt"
	"strings"
	"time"

	"github.com/d-fi/GoFi/internal/models"
	"github.com/d-fi/GoFi/logger"
	"github.com/d-fi/GoFi/types"
)

// SearchTrackOnDeezer searches for a Spotify track on Deezer
func SearchTrackOnDeezer(track *models.Track) (types.TrackType, error) {
	if track == nil {
		return types.TrackType{}, fmt.Errorf("cannot search for nil track")
	}

	// Try searching by ISRC first (most accurate)
	if track.ISRC != "" {
		logger.Debug("Searching for track by ISRC: %s", track.ISRC)
		searchResult, err := SearchMusic("isrc:"+track.ISRC, 1)
		if err == nil && len(searchResult.TRACK.Data) > 0 {
			trackID := fmt.Sprint(searchResult.TRACK.Data[0].SNG_ID)
			logger.Debug("Found track by ISRC match: %s", trackID)
			return GetTrackInfo(trackID)
		}
		logger.Debug("ISRC search failed or returned no results, falling back to metadata search")
	}

	// Clean up artist and track names for more accurate searching
	artistName := getMainArtistName(track.Artists)
	trackTitle := cleanupTitle(track.Title)

	// Build a search query
	query := fmt.Sprintf("artist:'%s' track:'%s'", artistName, trackTitle)
	logger.Debug("Searching for track with query: %s", query)

	searchResult, err := SearchMusic(query, 5)
	if err != nil {
		return types.TrackType{}, fmt.Errorf("failed to search for track: %v", err)
	}

	if len(searchResult.TRACK.Data) == 0 {
		return types.TrackType{}, fmt.Errorf("no matching tracks found on Deezer")
	}

	// Get the first result - could improve this by comparing durations, etc.
	trackID := fmt.Sprint(searchResult.TRACK.Data[0].SNG_ID)
	logger.Debug("Found potential track match: %s", trackID)
	
	return GetTrackInfo(trackID)
}

// SearchAlbumOnDeezer searches for a Spotify album on Deezer
func SearchAlbumOnDeezer(album *models.Album) (types.AlbumType, error) {
	if album == nil {
		return types.AlbumType{}, fmt.Errorf("cannot search for nil album")
	}

	// Try searching by UPC first (most accurate)
	if album.UPC != "" {
		logger.Debug("Searching for album by UPC: %s", album.UPC)
		searchResult, err := SearchMusic("upc:"+album.UPC, 1, "ALBUM")
		if err == nil && len(searchResult.ALBUM.Data) > 0 {
			albumID := fmt.Sprint(searchResult.ALBUM.Data[0].ALB_ID)
			logger.Debug("Found album by UPC match: %s", albumID)
			return GetAlbumInfo(albumID)
		}
		logger.Debug("UPC search failed or returned no results, falling back to metadata search")
	}

	// Clean up artist and album names for more accurate searching
	artistName := getMainArtistName(album.Artists)
	albumTitle := cleanupTitle(album.Title)

	// Build a search query
	query := fmt.Sprintf("artist:'%s' album:'%s'", artistName, albumTitle)
	logger.Debug("Searching for album with query: %s", query)

	searchResult, err := SearchMusic(query, 5, "ALBUM")
	if err != nil {
		return types.AlbumType{}, fmt.Errorf("failed to search for album: %v", err)
	}

	if len(searchResult.ALBUM.Data) == 0 {
		return types.AlbumType{}, fmt.Errorf("no matching albums found on Deezer")
	}

	// Get the first result
	albumID := fmt.Sprint(searchResult.ALBUM.Data[0].ALB_ID)
	logger.Debug("Found potential album match: %s", albumID)
	
	return GetAlbumInfo(albumID)
}

// MatchPlaylistTracks matches Spotify playlist tracks to Deezer tracks
func MatchPlaylistTracks(tracks []models.Track) ([]types.TrackType, error) {
	if len(tracks) == 0 {
		return nil, fmt.Errorf("cannot match empty track list")
	}

	logger.Debug("Starting to match %d Spotify tracks to Deezer", len(tracks))
	result := make([]types.TrackType, 0, len(tracks))
	failures := 0

	// Process each track, with rate limiting to avoid overwhelming the API
	for i, track := range tracks {
		// Add a small delay every few tracks to avoid rate limiting
		if i > 0 && i%5 == 0 {
			time.Sleep(500 * time.Millisecond)
		}

		logger.Debug("Processing track %d/%d: %s - %s", i+1, len(tracks), track.Title, getMainArtistName(track.Artists))
		deezerTrack, err := SearchTrackOnDeezer(&track)
		if err != nil {
			logger.Error("Failed to match track %s - %s: %v", track.Title, getMainArtistName(track.Artists), err)
			failures++
			continue
		}

		result = append(result, deezerTrack)
		logger.Debug("Successfully matched track %d/%d: %s to Deezer ID %s", 
			i+1, len(tracks), track.Title, deezerTrack.SNG_ID)
	}

	logger.Debug("Completed track matching: %d tracks matched, %d failed", len(result), failures)
	return result, nil
}

// Helper functions

// getMainArtistName returns the name of the main artist
func getMainArtistName(artists []models.Artist) string {
	if len(artists) == 0 {
		return ""
	}
	return artists[0].Name
}

// cleanupTitle removes common extra text from titles like "(Remastered)" or "[Live]"
func cleanupTitle(title string) string {
	// Remove content in parentheses, brackets, etc. that might differ between services
	cleaned := title

	// Using simple string replacements for common patterns
	if strings.Contains(cleaned, " (feat.") {
		cleaned = cleaned[:strings.Index(cleaned, " (feat.")]
	}
	if strings.Contains(cleaned, " (ft.") {
		cleaned = cleaned[:strings.Index(cleaned, " (ft.")]
	}
	if strings.Contains(cleaned, " (Remastered") {
		cleaned = cleaned[:strings.Index(cleaned, " (Remastered")]
	}
	if strings.Contains(cleaned, " [") {
		cleaned = cleaned[:strings.Index(cleaned, " [")]
	}
	if strings.Contains(cleaned, " - ") {
		cleaned = cleaned[:strings.Index(cleaned, " - ")]
	}

	return cleaned
}