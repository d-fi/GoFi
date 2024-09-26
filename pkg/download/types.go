package download

// UserData struct stores user license and streaming capabilities
type UserData struct {
	LicenseToken      string
	CanStreamLossless bool
	CanStreamHQ       bool
	Country           string
}

// TrackDownloadUrl represents the details of a track's download URL.
type TrackDownloadUrl struct {
	TrackUrl    string // The URL to download the track.
	IsEncrypted bool   // Indicates if the track URL points to an encrypted file.
	FileSize    int    // The size of the track file in bytes.
}

// TrackDownloadOptions contains all the details needed for downloading a track.
type TrackDownloadOptions struct {
	SngID     string // Song ID
	Quality   int    // Quality of the track (e.g., 128kbps, 320kbps, FLAC)
	Ext       string // File extension (e.g., mp3, flac)
	CoverSize int    // Cover size for album art
	SaveToDir string // Directory to save the downloaded track
}
