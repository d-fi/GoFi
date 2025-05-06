package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/d-fi/GoFi/internal/services/spotify"
	"github.com/d-fi/GoFi/logger"
	spotifyClient "github.com/zmb3/spotify/v2"
)

// Helper function to get authenticated client (handles prompting for auth)
// Returns nil client if auth fails or is not configured.
func getAuthenticatedSpotifyClient(ctx context.Context) (*spotifyClient.Client, *spotify.AuthService) {
	clientID := os.Getenv("SPOTIFY_CLIENT_ID")
	clientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		logger.Info("SPOTIFY_CLIENT_ID and SPOTIFY_CLIENT_SECRET environment variables not set. Spotify features disabled.")
		return nil, nil
	}
	cfg := spotify.Config{ClientID: clientID, ClientSecret: clientSecret}
	authService, err := spotify.NewAuthService(cfg)
	if err != nil {
		logger.Error("Error creating Spotify auth service: %v", err)
		return nil, nil
	}

	client, err := authService.GetClient(ctx) // Attempts to load token or refresh
	if err != nil {
		// This is not necessarily an error, just means we need to authenticate.
		// Logged as Info level. The calling function decides if it's fatal.
		logger.Info("Could not automatically retrieve Spotify token (may need initial auth): %v", err)
		fmt.Println("Could not retrieve Spotify token. Please run 'gofi auth spotify' first.")
		return nil, authService // Return authService even if client is nil
	}
	logger.Debug("Successfully obtained authenticated Spotify client.")
	return client, authService
}