// cmd/gofi/cmd/auth_spotify.go
package cmd

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/d-fi/GoFi/internal/services/spotify" // Corrected import path
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
			log.Fatal("Error: SPOTIFY_CLIENT_ID and SPOTIFY_CLIENT_SECRET environment variables must be set.")
			return // Redundant after Fatal, but good practice
		}

		cfg := spotify.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
		}

		authService, err := spotify.NewAuthService(cfg)
		if err != nil {
			log.Fatalf("Error creating Spotify auth service: %v", err)
		}

		log.Println("Starting Spotify authentication process...")
		// Use context.Background() for this CLI command execution context
		client, err := authService.StartAuthentication(context.Background())
		if err != nil {
			log.Fatalf("Spotify authentication failed: %v", err)
		}

		if client != nil {
			// Optionally, verify the client works by making a simple API call
			user, err := client.CurrentUser(context.Background())
			if err != nil {
				log.Printf("Warning: Could not verify authentication by fetching user: %v", err)
				fmt.Println("Authentication process completed, but verification failed. Token might be stored.")
			} else {
				fmt.Printf("\nSuccessfully authenticated with Spotify as %s (%s)!\n", user.DisplayName, user.ID)
				fmt.Println("Authentication token saved successfully.")
			}
		} else {
            // This case should ideally be caught by the error above, but added for completeness
            log.Fatal("Authentication process completed, but no valid client was returned.")
        }
	},
}

// 🔄 TODO: Register this command with your root command or an 'auth' group command.
// Example (in your root command setup, e.g., cmd/gofi/cmd/root.go or cmd/gofi/cmd/auth.go):
//
// import "github.com/spf13/cobra"
//
// var authCmd = &cobra.Command{
// 	 Use:   "auth",
// 	 Short: "Manage authentication for different services",
// }
//
// func init() {
//   // Add authCmd to rootCmd
//   rootCmd.AddCommand(authCmd)
//   // Add spotifyAuthCmd to authCmd
//   authCmd.AddCommand(spotifyAuthCmd)
//   // Or if no auth group, add directly to root:
//   // rootCmd.AddCommand(spotifyAuthCmd)
// }
//

func init() {
	// This function needs to exist to potentially register flags for this specific command later.
	// For now, we just need to ensure the command is added to its parent elsewhere.
	// If you have a central `auth` command:
	// authCmd.AddCommand(spotifyAuthCmd)
	// If adding directly to root command:
	// rootCmd.AddCommand(spotifyAuthCmd)
	// ---> You will need to uncomment and place the registration code in the appropriate file <---
	log.Println("Spotify auth command initialized (needs registration in root/auth command)") // Placeholder log
} 