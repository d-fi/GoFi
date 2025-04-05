package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/d-fi/GoFi/api"
	"github.com/d-fi/GoFi/download"
	"github.com/d-fi/GoFi/logger"
	"github.com/d-fi/GoFi/request"
	"github.com/d-fi/GoFi/types"
	"github.com/d-fi/GoFi/utils"
)

var (
	appConfig types.Config
)

func main() {
	// Set up command line flags
	var (
		arl        string
		configFile string
		logLevel   string
	)

	flag.StringVar(&arl, "arl", "", "Deezer ARL token (required for authentication)")
	flag.StringVar(&configFile, "config", "", "Path to config file (if not using command line args)")
	flag.StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	flag.Parse()

	// Set log level based on flag (debug environment variable is used in logger init)
	if logLevel == "debug" {
		os.Setenv("DEBUG", "true")
	}

	// Load configuration from file if specified
	appConfig = types.DefaultConfig()
	if configFile != "" {
		loadedConfig, err := loadConfigFromFile(configFile)
		if err != nil {
			fmt.Printf("Error loading config file: %v\n", err)
			os.Exit(1)
		}
		appConfig = loadedConfig
		logger.Debug("Loaded configuration from file: %s", configFile)
	}

	// Check for ARL in environment variable or config file if not provided via flag
	if arl == "" {
		arl = os.Getenv("DEEZER_ARL")
		if arl == "" && appConfig.Cookies.ARL != "" {
			arl = appConfig.Cookies.ARL
			logger.Debug("Using ARL from config file")
		}
	}

	// Validate ARL
	if arl == "" {
		fmt.Println("Error: Deezer ARL token is required.")
		fmt.Println("You can provide it using the -arl flag, set the DEEZER_ARL environment variable, or specify in config file.")
		fmt.Println("\nUsage: ./d-fi -arl=YOUR_ARL_TOKEN")
		os.Exit(1)
	}

	// Initialize Deezer API
	sessionID, err := request.InitDeezerAPI(arl)
	if err != nil {
		fmt.Printf("Failed to initialize Deezer API: %v\n", err)
		os.Exit(1)
	}

	logger.Debug("Successfully authenticated with Deezer (Session ID: %s)", sessionID)
	
	// Check for any arguments after flags
	args := flag.Args()
	if len(args) == 0 {
		printUsage()
		return
	}

	// Process the command
	command := args[0]
	switch command {
	case "search":
		if len(args) < 2 {
			fmt.Println("Error: Search query required")
			return
		}
		query := strings.Join(args[1:], " ")
		searchMusic(query)
	case "download":
		if len(args) < 3 {
			fmt.Println("Error: Usage: download <type> <id>")
			return
		}
		downloadType := args[1]
		id := args[2]
		
		switch downloadType {
		case "track":
			downloadTrack(id)
		case "album":
			downloadAlbum(id)
		case "playlist":
			downloadPlaylist(id)
		default:
			fmt.Printf("Error: Unknown download type: %s\n", downloadType)
			fmt.Println("Available types: track, album, playlist")
		}
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
	}
}

func loadConfigFromFile(configPath string) (types.Config, error) {
	config := types.DefaultConfig()
	
	// Read the config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return config, fmt.Errorf("failed to read config file: %v", err)
	}
	
	// Parse the JSON
	if err := json.Unmarshal(data, &config); err != nil {
		return config, fmt.Errorf("failed to parse config file: %v", err)
	}
	
	return config, nil
}

func printUsage() {
	fmt.Println("GoFi - Deezer music downloader")
	fmt.Println("\nUsage:")
	fmt.Println("  ./d-fi [options] command [arguments]")
	fmt.Println("\nOptions:")
	fmt.Println("  -arl string       Deezer ARL token (required for authentication)")
	fmt.Println("  -config string    Path to config file")
	fmt.Println("  -log-level string Log level (debug, info, warn, error) (default \"info\")")
	fmt.Println("\nCommands:")
	fmt.Println("  search <query>               Search for tracks, albums, or artists")
	fmt.Println("  download <type> <id>         Download track, album, or playlist")
	fmt.Println("    Types: track, album, playlist")
	fmt.Println("\nExamples:")
	fmt.Println("  ./d-fi -arl=YOUR_ARL search \"artist name\"")
	fmt.Println("  ./d-fi -arl=YOUR_ARL download track 123456789")
	fmt.Println("  ./d-fi -config=/path/to/config.json download album 123456789")
}

func searchMusic(query string) {
	fmt.Printf("Searching for: %s\n", query)
	
	results, err := api.SearchMusic(query, 10, "TRACK", "ALBUM", "ARTIST")
	if err != nil {
		fmt.Printf("Search failed: %v\n", err)
		return
	}

	// Display track results
	if len(results.TRACK.Data) > 0 {
		fmt.Println("\nTracks:")
		for i, track := range results.TRACK.Data {
			fmt.Printf("%d. %s - %s (ID: %s)\n", i+1, track.ART_NAME, track.SNG_TITLE, track.SNG_ID)
		}
	}

	// Display album results
	if len(results.ALBUM.Data) > 0 {
		fmt.Println("\nAlbums:")
		for i, album := range results.ALBUM.Data {
			fmt.Printf("%d. %s - %s (ID: %s)\n", i+1, album.ART_NAME, album.ALB_TITLE, album.ALB_ID)
		}
	}

	// Display artist results
	if len(results.ARTIST.Data) > 0 {
		fmt.Println("\nArtists:")
		for i, artist := range results.ARTIST.Data {
			fmt.Printf("%d. %s (ID: %s)\n", i+1, artist.ART_NAME, artist.ART_ID)
		}
	}
}

func downloadTrack(trackID string) {
	fmt.Println("Downloading track...")
	
	// Get the track info
	track, err := api.GetTrackInfo(trackID)
	if err != nil {
		fmt.Printf("Error: Failed to get track info: %v\n", err)
		return
	}
	
	// Create the download directory based on the config
	downloadDir := expandPathTemplate(appConfig.SaveLayout.Track, track, nil, "")
	dir := filepath.Dir(downloadDir)
	
	// Create the directory if it doesn't exist
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Printf("Error: Failed to create download directory: %v\n", err)
		return
	}
	
	// Determine quality based on track availability
	quality := 9 // FLAC by default
	
	// Progress tracking function
	progressFunc := func(progress float64, downloaded, total int64) {
		fmt.Printf("\rDownloading: %.1f%% (%s/%s)", progress, formatSize(downloaded), formatSize(total))
	}
	
	// Download the track
	options := download.DownloadTrackOptions{
		SngID:      trackID,
		Quality:    quality,
		CoverSize:  appConfig.CoverSize.Flac,
		SaveToDir:  dir,
		OnProgress: progressFunc,
	}
	
	filePath, err := download.DownloadTrack(options)
	if err != nil {
		fmt.Printf("\nError: Failed to download track: %v\n", err)
		return
	}
	
	fmt.Printf("\nTrack downloaded successfully: %s\n", filePath)
}

func downloadAlbum(albumID string) {
	fmt.Println("Downloading album...")
	
	// Get the album info
	album, err := api.GetAlbumInfo(albumID)
	if err != nil {
		fmt.Printf("Error: Failed to get album info: %v\n", err)
		return
	}
	
	// Get the tracks in the album
	tracks, err := api.GetAlbumTracks(albumID)
	if err != nil {
		fmt.Printf("Error: Failed to get album tracks: %v\n", err)
		return
	}
	
	fmt.Printf("Album: %s by %s\n", album.ALB_TITLE, album.ART_NAME)
	fmt.Printf("Total tracks: %d\n", len(tracks.Data))
	
	// Create a wait group to track download completion
	var wg sync.WaitGroup
	concurrency := appConfig.Concurrency
	semaphore := make(chan struct{}, concurrency)
	
	for i, track := range tracks.Data {
		wg.Add(1)
		semaphore <- struct{}{} // Acquire semaphore
		
		go func(i int, track types.TrackType) {
			defer wg.Done()
			defer func() { <-semaphore }() // Release semaphore
			
			fmt.Printf("[%d/%d] Downloading: %s - %s\n", i+1, len(tracks.Data), track.ART_NAME, track.SNG_TITLE)
			
			// Create the download directory based on the config
			downloadDir := expandPathTemplate(appConfig.SaveLayout.Album, track, &album, "")
			dir := filepath.Dir(downloadDir)
			
			// Create the directory if it doesn't exist
			if err := os.MkdirAll(dir, 0755); err != nil {
				fmt.Printf("Error: Failed to create download directory for track %s: %v\n", track.SNG_TITLE, err)
				return
			}
			
			// Determine quality based on track availability
			quality := 9 // FLAC by default
			
			// Download the track
			options := download.DownloadTrackOptions{
				SngID:     track.SNG_ID,
				Quality:   quality,
				CoverSize: appConfig.CoverSize.Flac,
				SaveToDir: dir,
			}
			
			filePath, err := download.DownloadTrack(options)
			if err != nil {
				fmt.Printf("Error: Failed to download track %s: %v\n", track.SNG_TITLE, err)
				return
			}
			
			fmt.Printf("Track downloaded: %s\n", filepath.Base(filePath))
		}(i, track)
	}
	
	wg.Wait()
	fmt.Println("Album download completed!")
}

func downloadPlaylist(playlistID string) {
	fmt.Println("Downloading playlist...")
	
	// Get the playlist info
	playlist, err := api.GetPlaylistInfo(playlistID)
	if err != nil {
		fmt.Printf("Error: Failed to get playlist info: %v\n", err)
		return
	}
	
	// Get the tracks in the playlist
	tracks, err := api.GetPlaylistTracks(playlistID)
	if err != nil {
		fmt.Printf("Error: Failed to get playlist tracks: %v\n", err)
		return
	}
	
	fmt.Printf("Playlist: %s\n", playlist.Title)
	fmt.Printf("Total tracks: %d\n", len(tracks.Data))
	
	// Create a wait group to track download completion
	var wg sync.WaitGroup
	concurrency := appConfig.Concurrency
	semaphore := make(chan struct{}, concurrency)
	
	for i, track := range tracks.Data {
		wg.Add(1)
		semaphore <- struct{}{} // Acquire semaphore
		
		go func(i int, track types.TrackType) {
			defer wg.Done()
			defer func() { <-semaphore }() // Release semaphore
			
			fmt.Printf("[%d/%d] Downloading: %s - %s\n", i+1, len(tracks.Data), track.ART_NAME, track.SNG_TITLE)
			
			// Create the download directory based on the config
			downloadDir := expandPathTemplate(appConfig.SaveLayout.Playlist, track, nil, playlist.Title)
			dir := filepath.Dir(downloadDir)
			
			// Create the directory if it doesn't exist
			if err := os.MkdirAll(dir, 0755); err != nil {
				fmt.Printf("Error: Failed to create download directory for track %s: %v\n", track.SNG_TITLE, err)
				return
			}
			
			// Determine quality based on track availability
			quality := 9 // FLAC by default
			
			// Download the track
			options := download.DownloadTrackOptions{
				SngID:     track.SNG_ID,
				Quality:   quality,
				CoverSize: appConfig.CoverSize.Flac,
				SaveToDir: dir,
			}
			
			filePath, err := download.DownloadTrack(options)
			if err != nil {
				fmt.Printf("Error: Failed to download track %s: %v\n", track.SNG_TITLE, err)
				return
			}
			
			fmt.Printf("Track downloaded: %s\n", filepath.Base(filePath))
		}(i, track)
	}
	
	wg.Wait()
	fmt.Println("Playlist download completed!")
}

// expandPathTemplate replaces placeholders in the template with actual values
func expandPathTemplate(template string, track types.TrackType, album *types.AlbumType, playlistTitle string) string {
	result := template
	
	// Track info
	result = strings.ReplaceAll(result, "{SNG_TITLE}", utils.SanitizeFileName(track.SNG_TITLE))
	result = strings.ReplaceAll(result, "{ART_NAME}", utils.SanitizeFileName(track.ART_NAME))
	result = strings.ReplaceAll(result, "{SNG_ID}", track.SNG_ID)
	
	// Track number (if available)
	trackNumberStr := ""
	if trackNum := int(track.TRACK_NUMBER); trackNum > 0 {
		trackNumberStr = fmt.Sprintf("%02d", trackNum)
	}
	result = strings.ReplaceAll(result, "{TRACK_NUMBER}", trackNumberStr)
	
	// Album info (if available)
	if album != nil {
		result = strings.ReplaceAll(result, "{ALB_TITLE}", utils.SanitizeFileName(album.ALB_TITLE))
		result = strings.ReplaceAll(result, "{ALB_ID}", album.ALB_ID)
	}
	
	// Playlist title (if available)
	if playlistTitle != "" {
		result = strings.ReplaceAll(result, "{TITLE}", utils.SanitizeFileName(playlistTitle))
	}
	
	return result
}

// formatSize formats a file size in bytes to a human-readable string
func formatSize(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	} else if size < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(size)/1024)
	} else if size < 1024*1024*1024 {
		return fmt.Sprintf("%.1f MB", float64(size)/1024/1024)
	} else {
		return fmt.Sprintf("%.1f GB", float64(size)/1024/1024/1024)
	}
}
