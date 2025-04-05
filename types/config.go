package types

// Config represents the application configuration
type Config struct {
	Concurrency int `json:"concurrency"`
	SaveLayout  struct {
		Track    string `json:"track"`
		Album    string `json:"album"`
		Artist   string `json:"artist"`
		Playlist string `json:"playlist"`
	} `json:"saveLayout"`
	Playlist struct {
		ResolveFullPath bool `json:"resolveFullPath"`
	} `json:"playlist"`
	TrackNumber     bool `json:"trackNumber"`
	FallbackTrack   bool `json:"fallbackTrack"`
	FallbackQuality bool `json:"fallbackQuality"`
	CoverSize       struct {
		Q128 int `json:"128"`
		Q320 int `json:"320"`
		Flac int `json:"flac"`
	} `json:"coverSize"`
	Cookies struct {
		ARL string `json:"arl"`
	} `json:"cookies"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() Config {
	config := Config{
		Concurrency:     1,
		TrackNumber:     false,
		FallbackTrack:   true,
		FallbackQuality: false,
	}

	// Default save layouts
	config.SaveLayout.Track = "./downloads/Tracks/{ART_NAME}/{ART_NAME} - {SNG_TITLE}"
	config.SaveLayout.Album = "./downloads/Albums/{ART_NAME}/{ALB_TITLE}/{TRACK_NUMBER} - {SNG_TITLE}"
	config.SaveLayout.Artist = "./downloads/Artists/{ART_NAME}/{SNG_TITLE}"
	config.SaveLayout.Playlist = "./downloads/Playlists/{TITLE}/{ART_NAME} - {SNG_TITLE}"

	// Default cover sizes
	config.CoverSize.Q128 = 500
	config.CoverSize.Q320 = 500
	config.CoverSize.Flac = 1000

	// Default playlist settings
	config.Playlist.ResolveFullPath = false

	return config
} 