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
	FileSize    int64  // The size of the track file in bytes.
}

// ProgressCallback defines a function type for tracking download progress.
type ProgressCallback func(progress float64, totalBytesRead, contentLength int64)

// DownloadTrackOptions contains all the details needed for downloading a track.
type DownloadTrackOptions struct {
	SngID      string                                          // The ID of the track to download.
	Quality    int                                             // The quality of the track (e.g., 1 for MP3_128, 3 for MP3_320, 9 for FLAC).
	CoverSize  int                                             // The size of the album cover in pixels.
	SaveToDir  string                                          // The directory where the track will be saved.
	OnProgress func(progress float64, downloaded, total int64) // The progress callback function.
}

// DownloadTrackToBufferOptions contains all the details needed for downloading a track to a buffer.
type DownloadTrackToBufferOptions struct {
	SngID     string // The ID of the track to download.
	Quality   int    // The quality of the track (e.g., 1 for MP3_128, 3 for MP3_320, 9 for FLAC).
	CoverSize int    // The size of the album cover in pixels.
}

// DownloadTrackWithoutMetadataOptions contains all the details needed for downloading a track without adding metadata.
type DownloadTrackWithoutMetadataOptions struct {
	SngID   string // The ID of the track to download.
	Quality int    // The quality of the track (e.g., 1 for MP3_128, 3 for MP3_320, 9 for FLAC).
}
