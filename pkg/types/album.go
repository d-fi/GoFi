package types

// AlbumTypeMinimal represents minimal information about an album, including artist and explicit content details.
type AlbumTypeMinimal struct {
	ALB_ID                 string       `json:"ALB_ID"`      // Album ID, e.g., '9188269'
	ALB_TITLE              string       `json:"ALB_TITLE"`   // Album title, e.g., 'The Days / Nights'
	ALB_PICTURE            string       `json:"ALB_PICTURE"` // Album picture hash, e.g., '6e58a99f59a150e9b4aefbeb2d6fc856'
	ARTISTS                []ArtistType `json:"ARTISTS"`     // List of artists
	AVAILABLE              bool         `json:"AVAILABLE"`   // Availability status
	VERSION                string       `json:"VERSION"`     // Version string, e.g., ''
	ART_ID                 string       `json:"ART_ID"`      // Artist ID, e.g., '293585'
	ART_NAME               string       `json:"ART_NAME"`    // Artist name, e.g., 'Avicii'
	EXPLICIT_ALBUM_CONTENT struct {     // Explicit content details
		EXPLICIT_LYRICS_STATUS int `json:"EXPLICIT_LYRICS_STATUS"` // Explicit lyrics status, e.g., 1
		EXPLICIT_COVER_STATUS  int `json:"EXPLICIT_COVER_STATUS"`  // Explicit cover status, e.g., 2
	} `json:"EXPLICIT_ALBUM_CONTENT"`
	PHYSICAL_RELEASE_DATE string `json:"PHYSICAL_RELEASE_DATE"` // Physical release date
	TYPE                  string `json:"TYPE"`                  // Album type, e.g., '0'
	ARTIST_IS_DUMMY       bool   `json:"ARTIST_IS_DUMMY"`       // Indicates if the artist is a dummy
	NUMBER_TRACK          string `json:"NUMBER_TRACK"`          // Number of tracks, e.g., '1'
	TYPE_INTERNAL         string `json:"__TYPE__"`              // Internal type, e.g., 'album'
}

// AlbumType represents detailed information about an album including contributors and release dates.
type AlbumType struct {
	ALB_CONTRIBUTORS struct { // Contributors to the album
		MainArtist []string `json:"main_artist"` // Main artist, e.g., ['Avicii']
	} `json:"ALB_CONTRIBUTORS"`
	ALB_ID                 string   `json:"ALB_ID"`      // Album ID, e.g., '9188269'
	ALB_PICTURE            string   `json:"ALB_PICTURE"` // Album picture hash, e.g., '6e58a99f59a150e9b4aefbeb2d6fc856'
	EXPLICIT_ALBUM_CONTENT struct { // Explicit content details
		EXPLICIT_LYRICS_STATUS int `json:"EXPLICIT_LYRICS_STATUS"` // Explicit lyrics status, e.g., 0
		EXPLICIT_COVER_STATUS  int `json:"EXPLICIT_COVER_STATUS"`  // Explicit cover status, e.g., 0
	} `json:"EXPLICIT_ALBUM_CONTENT"`
	ALB_TITLE             string       `json:"ALB_TITLE"`                       // Album title, e.g., 'The Days / Nights'
	ARTISTS               []ArtistType `json:"ARTISTS"`                         // List of artists
	ART_ID                string       `json:"ART_ID"`                          // Artist ID, e.g., '293585'
	ART_NAME              string       `json:"ART_NAME"`                        // Artist name, e.g., 'Avicii'
	ARTIST_IS_DUMMY       bool         `json:"ARTIST_IS_DUMMY"`                 // Indicates if the artist is a dummy
	DIGITAL_RELEASE_DATE  string       `json:"DIGITAL_RELEASE_DATE"`            // Digital release date, e.g., '2014-12-01'
	EXPLICIT_LYRICS       *string      `json:"EXPLICIT_LYRICS,omitempty"`       // Optional explicit lyrics status
	NB_FAN                int          `json:"NB_FAN"`                          // Number of fans, e.g., 36285
	NUMBER_DISK           string       `json:"NUMBER_DISK"`                     // Number of disks, e.g., '1'
	NUMBER_TRACK          string       `json:"NUMBER_TRACK"`                    // Number of tracks, e.g., '1'
	PHYSICAL_RELEASE_DATE *string      `json:"PHYSICAL_RELEASE_DATE,omitempty"` // Optional physical release date
	PRODUCER_LINE         string       `json:"PRODUCER_LINE"`                   // Producer line, e.g., '℗ 2014 Avicii Music AB'
	PROVIDER_ID           string       `json:"PROVIDER_ID"`                     // Provider ID, e.g., '427'
	RANK                  string       `json:"RANK"`                            // Rank, e.g., '601128'
	RANK_ART              string       `json:"RANK_ART"`                        // Artist rank, e.g., '861905'
	STATUS                string       `json:"STATUS"`                          // Status, e.g., '1'
	TYPE                  string       `json:"TYPE"`                            // Type, e.g., '1'
	UPC                   string       `json:"UPC"`                             // UPC, e.g., '602547151544'
	TYPE_INTERNAL         string       `json:"__TYPE__"`                        // Internal type, e.g., 'album'
}

// AlbumTracksType represents track details in an album including count and filtering information.
type AlbumTracksType struct {
	Data          []TrackType `json:"data"`                     // List of track data
	Count         int         `json:"count"`                    // Count of tracks
	Total         int         `json:"total"`                    // Total tracks in album
	FilteredCount int         `json:"filtered_count"`           // Filtered count
	FilteredItems []int       `json:"filtered_items,omitempty"` // Optional filtered item IDs
	Next          int         `json:"next,omitempty"`           // Next page number if available
}

// TrackDataPublicApi represents a track's public information from the API.
type TrackDataPublicApi struct {
	ID                    int              `json:"id"`                      // Track ID, e.g., 3135556
	Readable              bool             `json:"readable"`                // Readable status
	Title                 string           `json:"title"`                   // Track title, e.g., 'Harder, Better, Faster, Stronger'
	TitleShort            string           `json:"title_short"`             // Short title, e.g., 'Harder, Better, Faster, Stronger'
	TitleVersion          string           `json:"title_version,omitempty"` // Optional version title
	Link                  string           `json:"link"`                    // Deezer link, e.g., 'https://www.deezer.com/track/3135556'
	Duration              int              `json:"duration"`                // Duration in seconds, e.g., 224
	Rank                  int              `json:"rank"`                    // Track rank, e.g., 956167
	ExplicitLyrics        bool             `json:"explicit_lyrics"`         // Explicit lyrics status
	ExplicitContentLyrics int              `json:"explicit_content_lyrics"` // Explicit content lyrics status
	ExplicitContentCover  int              `json:"explicit_content_cover"`  // Explicit content cover status
	Preview               string           `json:"preview"`                 // Preview link
	MD5Image              string           `json:"md5_image"`               // MD5 hash of the image
	Artist                ArtistDataPublic `json:"artist"`                  // Artist details
	Type                  string           `json:"type"`                    // Type, e.g., 'track'
}

// ArtistDataPublic represents public information about an artist.
type ArtistDataPublic struct {
	ID        int    `json:"id"`        // Artist ID, e.g., 27
	Name      string `json:"name"`      // Artist name, e.g., 'Daft Punk'
	Tracklist string `json:"tracklist"` // Artist tracklist link, e.g., 'https://api.deezer.com/artist/27/top?limit=50'
	Type      string `json:"type"`      // Type, e.g., 'artist'
}

// GenreTypePublicApi represents genre information from the public API.
type GenreTypePublicApi struct {
	ID      int    `json:"id"`      // Genre ID, e.g., 113
	Name    string `json:"name"`    // Genre name, e.g., 'Dance'
	Picture string `json:"picture"` // Genre picture URL
	Type    string `json:"type"`    // Type, e.g., 'genre'
}

// AlbumTypePublicApi represents detailed album information from the public API.
type AlbumTypePublicApi struct {
	ID                    int                     `json:"id"`                      // Album ID, e.g., 302127
	Title                 string                  `json:"title"`                   // Album title, e.g., 'Discovery'
	UPC                   string                  `json:"upc"`                     // UPC code, e.g., '724384960650'
	Link                  string                  `json:"link"`                    // Deezer link, e.g., 'https://www.deezer.com/album/302127'
	Share                 string                  `json:"share"`                   // Share link
	Cover                 string                  `json:"cover"`                   // Cover image URL
	CoverSmall            string                  `json:"cover_small"`             // Small cover image URL
	CoverMedium           string                  `json:"cover_medium"`            // Medium cover image URL
	CoverBig              string                  `json:"cover_big"`               // Big cover image URL
	CoverXL               string                  `json:"cover_xl"`                // Extra-large cover image URL
	MD5Image              string                  `json:"md5_image"`               // MD5 image hash
	GenreID               int                     `json:"genre_id"`                // Genre ID, e.g., 113
	Genres                GenreTypePublicApiList  `json:"genres"`                  // List of genres
	Label                 string                  `json:"label"`                   // Label, e.g., 'Parlophone (France)'
	NbTracks              int                     `json:"nb_tracks"`               // Number of tracks
	Duration              int                     `json:"duration"`                // Duration in seconds
	Fans                  int                     `json:"fans"`                    // Number of fans
	Rating                int                     `json:"rating"`                  // Album rating
	ReleaseDate           string                  `json:"release_date"`            // Release date, e.g., '2001-03-07'
	RecordType            string                  `json:"record_type"`             // Record type, e.g., 'album'
	Available             bool                    `json:"available"`               // Availability status
	Tracklist             string                  `json:"tracklist"`               // Tracklist URL
	ExplicitLyrics        bool                    `json:"explicit_lyrics"`         // Explicit lyrics status
	ExplicitContentLyrics int                     `json:"explicit_content_lyrics"` // Explicit content lyrics status
	ExplicitContentCover  int                     `json:"explicit_content_cover"`  // Explicit content cover status
	Contributors          []ContributorsPublicApi `json:"contributors"`            // List of contributors
	Artist                ContributorsPublicApi   `json:"artist"`                  // Artist details
	Type                  string                  `json:"type"`                    // Type, e.g., 'album'
	Tracks                TrackDataPublicApiList  `json:"tracks"`                  // List of track data
}

// GenreTypePublicApiList represents a list of genres from the public API.
type GenreTypePublicApiList struct {
	Data []GenreTypePublicApi `json:"data"` // Array of genre details
}

// ContributorsPublicApi represents information about contributors from the public API.
type ContributorsPublicApi struct {
	ID            int    `json:"id"`             // Contributor ID
	Name          string `json:"name"`           // Contributor name
	Link          string `json:"link"`           // Link to contributor profile
	Share         string `json:"share"`          // Shareable link
	Picture       string `json:"picture"`        // Picture URL
	PictureSmall  string `json:"picture_small"`  // Small picture URL
	PictureMedium string `json:"picture_medium"` // Medium picture URL
	PictureBig    string `json:"picture_big"`    // Big picture URL
	PictureXL     string `json:"picture_xl"`     // Extra-large picture URL
	Radio         bool   `json:"radio"`          // Radio availability
	Tracklist     string `json:"tracklist"`      // Tracklist link
	Type          string `json:"type"`           // Type, e.g., 'artist'
}

// TrackDataPublicApiList represents a list of track data from the public API.
type TrackDataPublicApiList struct {
	Data []TrackDataPublicApi `json:"data"` // Array of track data
}
