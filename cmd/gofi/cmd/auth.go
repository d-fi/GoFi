package cmd

import (
	"github.com/spf13/cobra"
)

// authCmd represents the auth command
var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate with music services",
	Long:  `Authenticate with supported music services like Spotify and Deezer.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	// Add sub-commands to auth
	authCmd.AddCommand(spotifyAuthCmd)
}