package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// downloadCmd represents the download command
var downloadCmd = &cobra.Command{
	Use:   "download [url]",
	Short: "Download music from a URL",
	Long: `Download music from Deezer or Spotify URL.
Supports tracks, albums, and playlists from both services.

For Spotify URLs, content will be searched on Deezer and downloaded.
For Deezer URLs, content will be downloaded directly.

Examples:
  gofi download https://open.spotify.com/track/2YarjDYjBJuH63dUIh9OWv
  gofi download https://open.spotify.com/album/1DFixLWuPkv3KT3TnV35m3
  gofi download https://open.spotify.com/playlist/37i9dQZF1DXcBWIGoYBM5M
  gofi download https://www.deezer.com/track/1234567
  gofi download https://www.deezer.com/album/1234567
  gofi download https://www.deezer.com/playlist/1234567890`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		url := args[0]
		fmt.Printf("Processing URL: %s\n", url)
		
		err := downloadHandler(url, downloadPath, quality)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	},
}