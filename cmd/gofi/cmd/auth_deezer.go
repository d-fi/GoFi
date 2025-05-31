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

		// Save to .env file
		if err := auth.SaveARLToEnv(arl); err != nil {
			color.Yellow("⚠️  Failed to save ARL to .env file: %v", err)
			fmt.Println("You can manually set the DEEZER_ARL environment variable.")
		} else {
			color.Green("✅ ARL token saved to .env file")
		}

		// Also set it in the current environment
		os.Setenv("DEEZER_ARL", arl)

		color.Green("\n✅ Successfully authenticated with Deezer!")
		fmt.Println("You can now download music from Deezer.")
		
		// Show a preview of the ARL (masked for security)
		if len(arl) > 20 {
			fmt.Printf("\nARL Token: %s...%s\n", arl[:10], arl[len(arl)-10:])
		}
	},
}

func init() {
	authCmd.AddCommand(authDeezerCmd)
}