// cmd/gofi/cmd/auth_spotify.go
package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/d-fi/GoFi/internal/services/spotify" // Corrected import path
	"github.com/d-fi/GoFi/logger"
	"github.com/spf13/cobra"
)

// spotifyAuthCmd represents the command to authenticate with Spotify
var spotifyAuthCmd = &cobra.Command{
	Use:   "spotify",
	Short: "Authenticate GoFi with your Spotify account",
	Long: `Starts the OAuth2 flow to authorize GoFi to access your Spotify account.
You will be prompted to open a URL in your browser to grant permission.
Requires SPOTIFY_CLIENT_ID and SPOTIFY_CLIENT_SECRET environment variables to be set.`,
	Run: func(cmd *cobra.Command, args []string) {
		clientID := os.Getenv("SPOTIFY_CLIENT_ID")
		clientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")

		if clientID == "" || clientSecret == "" {
			logger.Fatal("Error: SPOTIFY_CLIENT_ID and SPOTIFY_CLIENT_SECRET environment variables must be set.")
			return // Redundant after Fatal, but good practice
		}

		cfg := spotify.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
		}

		authService, err := spotify.NewAuthService(cfg)
		if err != nil {
			logger.Fatal("Error creating Spotify auth service: %v", err)
		}

		logger.Info("Starting Spotify authentication process...")
		// Use context.Background() for this CLI command execution context
		client, err := authService.StartAuthentication(context.Background())
		if err != nil {
			logger.Fatal("Spotify authentication failed: %v", err)
		}

		if client != nil {
			// Optionally, verify the client works by making a simple API call
			user, err := client.CurrentUser(context.Background())
			if err != nil {
				logger.Warn("Could not verify authentication by fetching user: %v", err)
				fmt.Println("Authentication process completed, but verification failed. Token might be stored.")
			} else {
				fmt.Printf("\nSuccessfully authenticated with Spotify as %s (%s)!\n", user.DisplayName, user.ID)
				fmt.Println("Authentication token saved successfully.")
			}
		} else {
			// This case should ideally be caught by the error above, but added for completeness
			logger.Fatal("Authentication process completed, but no valid client was returned.")
		}
	},
}

// init function is already called from auth.go
// No need to add logging or registration here
func init() {
	// Empty init function - registration handled in auth.go
}