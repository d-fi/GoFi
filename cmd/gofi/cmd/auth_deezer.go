package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/d-fi/GoFi/internal/auth"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var authDeezerCmd = &cobra.Command{
	Use:   "deezer",
	Short: "Authenticate with Deezer using browser cookies",
	Long: `Authenticate with Deezer by reading the ARL cookie from your browser.
This command will automatically check Chrome, Firefox, Edge, Arc, and Safari (on macOS)
for the Deezer ARL cookie and save it to your environment.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Try to get ARL from browser cookies
		fmt.Println("🔍 Searching for Deezer ARL cookie in your browsers...")
		
		arl, err := auth.GetARLFromAnyBrowser()
		if err != nil {
			color.Red("❌ Failed to find Deezer ARL cookie: %v", err)
			fmt.Println("\nPlease make sure you are logged into Deezer in one of the following browsers:")
			fmt.Println("  • Chrome")
			fmt.Println("  • Firefox")
			fmt.Println("  • Edge")
			fmt.Println("  • Arc")
			if runtime.GOOS == "darwin" {
				fmt.Println("  • Safari")
			}
			fmt.Println("\nAlternatively, you can set the DEEZER_ARL environment variable manually.")
			os.Exit(1)
		}

		// Validate the ARL token
		if err := auth.ValidateARLToken(arl); err != nil {
			color.Red("❌ Invalid ARL token: %v", err)
			os.Exit(1)
		}

		// Clean the ARL token before using it
		cleanARL := ""
		for _, r := range arl {
			if r >= 32 && r <= 126 {
				cleanARL += string(r)
			}
		}
		
		// Save to .env file
		if err := auth.SaveARLToEnv(cleanARL); err != nil {
			color.Yellow("⚠️  Failed to save ARL to .env file: %v", err)
			fmt.Println("You can manually set the DEEZER_ARL environment variable.")
		} else {
			color.Green("✅ ARL token saved to .env file")
		}

		// Also set it in the current environment
		os.Setenv("DEEZER_ARL", cleanARL)

		color.Green("\n✅ Successfully authenticated with Deezer!")
		fmt.Println("You can now download music from Deezer.")
		
		// Show a preview of the ARL (masked for security)
		if len(arl) > 20 {
			// Clean the display - only show printable characters
			start := ""
			end := ""
			
			// Get first 10 printable characters
			for i, r := range arl {
				if r >= 32 && r <= 126 {
					start += string(r)
					if len(start) >= 10 {
						break
					}
				}
				if i > 50 { // Don't search too far
					break
				}
			}
			
			// Get last 10 characters (usually clean)
			if len(arl) >= 10 {
				end = arl[len(arl)-10:]
			}
			
			fmt.Printf("\nARL Token: %s...%s\n", start, end)
		}
	},
}

func init() {
	authCmd.AddCommand(authDeezerCmd)
}