// internal/models/models.go
package models

import "time" // Added for potential timestamp fields

// SourceType indicates the origin of the data (Spotify, Deezer, etc.)
type SourceType string

const (
	SourceSpotify SourceType = "Spotify"
	SourceDeezer  SourceType = "Deezer"
	SourceTidal   SourceType = "Tidal"
	SourceUnknown SourceType = "Unknown"
	// Add other sources as needed
)

// Image represents an image URL with dimensions.
type Image struct {
	URL    string `json:"url"`
	Height int    `json:"height,omitempty"`
	Width  int    `json:"width,omitempty"`
}

// Artist represents a music artist.
type Artist struct {
	ID     string     `json:"id,omitempty"` // Service-specific ID
	Source SourceType `json:"source,omitempty"`
	Name   string     `json:"name"`
	Images []Image    `json:"images,omitempty"`
}

// Album represents a music album.
type Album struct {
	ID          string     `json:"id,omitempty"` // Service-specific ID (often same as SpotifyID for Spotify)
	SpotifyID   string     `json:"spotify_id,omitempty"` // Explicit Spotify ID
	Source      SourceType `json:"source,omitempty"`
	Title       string     `json:"title"`
	Artists     []Artist   `json:"artists,omitempty"`
	Images      []Image    `json:"images,omitempty"`
	ReleaseDate string     `json:"release_date,omitempty"` // Consider time.Time if precision is needed
	Label       string     `json:"label,omitempty"` // Record label
	TotalTracks int        `json:"total_tracks,omitempty"`
	UPC         string     `json:"upc,omitempty"`    // Universal Product Code
	Genres      []string   `json:"genres,omitempty"` // Added Genres
	AlbumType   string     `json:"album_type,omitempty"` // Added AlbumType (e.g., "album", "single")
}

// Track represents a music track.
type Track struct {
	ID           string     `json:"id,omitempty"` // Service-specific ID
	SpotifyID    string     `json:"spotify_id,omitempty"` // Explicit Spotify ID
	Source       SourceType `json:"source,omitempty"`
	Title        string     `json:"title"`
	Artists      []Artist   `json:"artists,omitempty"`
	Album        *Album     `json:"album,omitempty"` // Changed from string to *Album
	DurationMs   int        `json:"duration_ms,omitempty"`
	TrackNumber  int        `json:"track_number,omitempty"`
	DiscNumber   int        `json:"disc_number,omitempty"`
	Explicit     bool       `json:"explicit,omitempty"`
	ISRC         string     `json:"isrc,omitempty"` // International Standard Recording Code
	Images       []Image    `json:"images,omitempty"` // Often inherited from Album
	ReleaseDate  string     `json:"release_date,omitempty"` // Often inherited from Album
	AddedAt      *time.Time `json:"added_at,omitempty"`     // For playlists
	DownloadURL  string     `json:"download_url,omitempty"` // Potential field for download link
	FileSize     int64      `json:"file_size,omitempty"`    // Potential field for file size
	AudioQuality string     `json:"audio_quality,omitempty"` // Potential field for quality info
}

// Playlist represents a music playlist.
type Playlist struct {
	ID          string   `json:"id,omitempty"` // Service-specific ID
	SpotifyID   string   `json:"spotify_id,omitempty"` // Explicit Spotify ID
	Source      SourceType `json:"source,omitempty"`
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	OwnerName   string   `json:"owner_name,omitempty"`
	Public      bool     `json:"public,omitempty"`
	TotalTracks int      `json:"total_tracks,omitempty"`
	Images      []Image  `json:"images,omitempty"`
	Tracks      []Track  `json:"tracks,omitempty"` // Populated after fetching details
} 