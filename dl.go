package main

import (
	"context"
	"log"

	"github.com/d-fi/GoFi/download"
)

func main() {
	options := download.DownloadTrackWithoutMetadataOptions{
		SngID:   "12345678", // Replace with the actual track ID
		Quality: 9,          // Quality level, e.g., 1 for MP3_128, 3 for MP3_320, 9 for FLAC
	}

	// Call the function to download the track without metadata
	trackData, err := download.DownloadTrackWithoutMetadata(context.Background(), options)
	if err != nil {
		log.Fatalf("Failed to download track: %v", err)
	}

	// The track data is now available in the trackData variable
}
