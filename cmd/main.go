package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
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
			fmt.Println("Error: Usage: download <type> <name|id>")
			return
		}
		downloadType := args[1]
		nameOrID := strings.Join(args[2:], " ")
		
		switch downloadType {
		case "track":
			downloadTrackByNameOrID(nameOrID)
		case "album":
			downloadAlbumByNameOrID(nameOrID)
		case "playlist":
			downloadPlaylistByNameOrID(nameOrID)
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
	fmt.Println("  download <type> <name|id>    Download track, album, or playlist by name or ID")
	fmt.Println("    Types: track, album, playlist")
	fmt.Println("\nExamples:")
	fmt.Println("  ./d-fi -arl=YOUR_ARL search \"daft punk\"")
	fmt.Println("  ./d-fi -arl=YOUR_ARL download track \"Harder Better Faster Stronger\"")
	fmt.Println("  ./d-fi -arl=YOUR_ARL download track 3135556")
	fmt.Println("  ./d-fi -config=config.json download album \"Discovery\"")
	fmt.Println("  ./d-fi -config=config.json download album 302127")
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

func downloadTrackByNameOrID(nameOrID string) {
	// Check if the input is an ID (all digits)
	isID := true
	for _, char := range nameOrID {
		if char < '0' || char > '9' {
			isID = false
			break
		}
	}

	var trackID string
	if isID {
		trackID = nameOrID
		fmt.Printf("Downloading track with ID: %s\n", trackID)
	} else {
		// Search for the track by name
		fmt.Printf("Searching for track: %s\n", nameOrID)
		results, err := api.SearchMusic(nameOrID, 5, "TRACK")
		if err != nil {
			fmt.Printf("Error: Failed to search for track: %v\n", err)
			return
		}

		if len(results.TRACK.Data) == 0 {
			fmt.Println("Error: No tracks found with that name")
			return
		}

		// Display the found tracks and ask the user to select one
		fmt.Println("Found tracks:")
		for i, track := range results.TRACK.Data {
			fmt.Printf("%d. %s - %s (ID: %s)\n", i+1, track.ART_NAME, track.SNG_TITLE, track.SNG_ID)
		}

		if len(results.TRACK.Data) > 1 {
			fmt.Print("\nSelect a track (1-5) or press Enter for the first result: ")
			var choice string
			fmt.Scanln(&choice)

			if choice == "" {
				trackID = results.TRACK.Data[0].SNG_ID
			} else {
				index, err := strconv.Atoi(choice)
				if err != nil || index < 1 || index > len(results.TRACK.Data) {
					fmt.Println("Invalid selection, using the first result")
					trackID = results.TRACK.Data[0].SNG_ID
				} else {
					trackID = results.TRACK.Data[index-1].SNG_ID
				}
			}
		} else {
			trackID = results.TRACK.Data[0].SNG_ID
		}
	}

	downloadTrack(trackID)
}

func downloadAlbumByNameOrID(nameOrID string) {
	// Check if the input is an ID (all digits)
	isID := true
	for _, char := range nameOrID {
		if char < '0' || char > '9' {
			isID = false
			break
		}
	}

	var albumID string
	if isID {
		albumID = nameOrID
		fmt.Printf("Downloading album with ID: %s\n", albumID)
	} else {
		// Search for the album by name
		fmt.Printf("Searching for album: %s\n", nameOrID)
		results, err := api.SearchMusic(nameOrID, 5, "ALBUM")
		if err != nil {
			fmt.Printf("Error: Failed to search for album: %v\n", err)
			return
		}

		if len(results.ALBUM.Data) == 0 {
			fmt.Println("Error: No albums found with that name")
			return
		}

		// Display the found albums and ask the user to select one
		fmt.Println("Found albums:")
		for i, album := range results.ALBUM.Data {
			fmt.Printf("%d. %s - %s (ID: %s)\n", i+1, album.ART_NAME, album.ALB_TITLE, album.ALB_ID)
		}

		if len(results.ALBUM.Data) > 1 {
			fmt.Print("\nSelect an album (1-5) or press Enter for the first result: ")
			var choice string
			fmt.Scanln(&choice)

			if choice == "" {
				albumID = results.ALBUM.Data[0].ALB_ID
			} else {
				index, err := strconv.Atoi(choice)
				if err != nil || index < 1 || index > len(results.ALBUM.Data) {
					fmt.Println("Invalid selection, using the first result")
					albumID = results.ALBUM.Data[0].ALB_ID
				} else {
					albumID = results.ALBUM.Data[index-1].ALB_ID
				}
			}
		} else {
			albumID = results.ALBUM.Data[0].ALB_ID
		}
	}

	downloadAlbum(albumID)
}

func downloadPlaylistByNameOrID(nameOrID string) {
	// Check if the input is an ID (all digits)
	isID := true
	for _, char := range nameOrID {
		if char < '0' || char > '9' {
			isID = false
			break
		}
	}

	var playlistID string
	if isID {
		playlistID = nameOrID
		fmt.Printf("Downloading playlist with ID: %s\n", playlistID)
	} else {
		// Search for the playlist by name
		fmt.Printf("Searching for playlist: %s\n", nameOrID)
		results, err := api.SearchMusic(nameOrID, 5, "PLAYLIST")
		if err != nil {
			fmt.Printf("Error: Failed to search for playlist: %v\n", err)
			return
		}

		if len(results.PLAYLIST.Data) == 0 {
			fmt.Println("Error: No playlists found with that name")
			return
		}

		// Display the found playlists and ask the user to select one
		fmt.Println("Found playlists:")
		for i, playlist := range results.PLAYLIST.Data {
			fmt.Printf("%d. %s (ID: %s)\n", i+1, playlist.Title, playlist.PlaylistID)
		}

		if len(results.PLAYLIST.Data) > 1 {
			fmt.Print("\nSelect a playlist (1-5) or press Enter for the first result: ")
			var choice string
			fmt.Scanln(&choice)

			if choice == "" {
				playlistID = results.PLAYLIST.Data[0].PlaylistID
			} else {
				index, err := strconv.Atoi(choice)
				if err != nil || index < 1 || index > len(results.PLAYLIST.Data) {
					fmt.Println("Invalid selection, using the first result")
					playlistID = results.PLAYLIST.Data[0].PlaylistID
				} else {
					playlistID = results.PLAYLIST.Data[index-1].PlaylistID
				}
			}
		} else {
			playlistID = results.PLAYLIST.Data[0].PlaylistID
		}
	}

	downloadPlaylist(playlistID)
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
