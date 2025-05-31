package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/d-fi/GoFi/logger"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

var (
	downloadPath string
	quality      int
	logLevel     string
	version      = "dev" // Set by build flags
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gofi",
	Short: "GoFi is a music download tool for Deezer with Spotify integration",
	Long: `GoFi is a music download tool written in Go that allows you to search
and download music from Deezer. It now supports Spotify integration for finding
tracks on Deezer using Spotify URLs.

You can download tracks, albums, and playlists in different qualities.`,
	Version: version,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Set up logging
		level, err := zerolog.ParseLevel(logLevel)
		if err != nil {
			fmt.Printf("Warning: Invalid log level '%s', defaulting to 'info'\n", logLevel)
			level = zerolog.InfoLevel
		}
		logger.SetLogLevel(level)
		
		// Ensure download path exists
		if downloadPath != "" {
			err := os.MkdirAll(downloadPath, 0755)
			if err != nil {
				fmt.Printf("Error creating download directory: %v\n", err)
				os.Exit(1)
			}
			
			// Convert to absolute path
			absPath, err := filepath.Abs(downloadPath)
			if err != nil {
				fmt.Printf("Error resolving download path: %v\n", err)
				os.Exit(1)
			}
			downloadPath = absPath
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// Load environment variables from .env file if it exists
	loadEnvFile()
	
	// Persistent flags that are global across all commands
	rootCmd.PersistentFlags().StringVarP(&downloadPath, "output", "o", "./downloads", "Directory to save downloaded files")
	rootCmd.PersistentFlags().IntVarP(&quality, "quality", "q", 3, "Audio quality (1=128kbps MP3, 3=320kbps MP3, 9=FLAC)")
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", "info", "Log level (debug, info, warn, error)")
	
	// Add subcommands
	rootCmd.AddCommand(authCmd)
	rootCmd.AddCommand(downloadCmd)
}

// loadEnvFile loads environment variables from .env file
func loadEnvFile() {
	// Try to open .env file
	data, err := os.ReadFile(".env")
	if err != nil {
		// .env file doesn't exist or can't be read, which is fine
		logger.Debug("No .env file found or unable to read it: %v", err)
		return
	}

	logger.Debug("Loading environment variables from .env file")

	// Parse and set environment variables
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		// Skip empty lines and comments
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split key and value
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			// Skip lines that don't have a key=value format
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		
		// Remove quotes if present
		if len(value) > 1 && (value[0] == '"' || value[0] == '\'') && value[0] == value[len(value)-1] {
			value = value[1 : len(value)-1]
		}
		
		// Set environment variable
		os.Setenv(key, value)
		logger.Debug("Set environment variable: %s", key)
	}
}