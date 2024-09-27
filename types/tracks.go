package types

import (
	"encoding/json"
	"fmt"
)

// MediaType represents media details, including the type and URL.
type MediaType struct {
	TYPE string `json:"TYPE"` // 'preview'
	HREF string `json:"HREF"` // 'https://cdns-preview-d.dzcdn.net/stream/c-deda7fa9316d9e9e880d2c6207e92260-8.mp3'
}

// LyricsSync represents the synchronized lyrics with timestamps.
type LyricsSync struct {
	LrcTimestamp string      `json:"lrc_timestamp"` // '[00:03.58]'
	Milliseconds string      `json:"milliseconds"`  // '3580'
	Duration     StringOrInt `json:"duration"`      // '8660'
	Line         string      `json:"line"`          // "Hey brother! There's an endless road to rediscover"
}

// LyricsType represents lyrics information, including text and synchronized data.
type LyricsType struct {
	LYRICS_ID         *string      `json:"LYRICS_ID,omitempty"`         // '2310758'
	LYRICS_SYNC_JSON  []LyricsSync `json:"LYRICS_SYNC_JSON,omitempty"`  // Array of synchronized lyrics
	LYRICS_TEXT       string       `json:"LYRICS_TEXT"`                 // Lyrics text
	LYRICS_COPYRIGHTS *string      `json:"LYRICS_COPYRIGHTS,omitempty"` // Optional lyrics copyrights
	LYRICS_WRITERS    *string      `json:"LYRICS_WRITERS,omitempty"`    // Optional lyrics writers
}

// ExplicitTrackContent represents explicit content status for lyrics and covers.
type ExplicitTrackContent struct {
	ExplicitLyricsStatus int `json:"EXPLICIT_LYRICS_STATUS"` // 0
	ExplicitCoverStatus  int `json:"EXPLICIT_COVER_STATUS"`  // 0
}

// SongContributors represents contributors to the song such as artists, authors, and more.
type SongContributors struct {
	MainArtist     []string `json:"main_artist,omitempty"`    // Main artist
	Author         []string `json:"author,omitempty"`         // Song authors
	Composer       []string `json:"composer,omitempty"`       // Composers
	MusicPublisher []string `json:"musicpublisher,omitempty"` // Music publishers
	Producer       []string `json:"producer,omitempty"`       // Producers
	Publisher      []string `json:"publisher"`                // Publishers
	Engineer       []string `json:"engineer,omitempty"`       // Engineers
	Writer         []string `json:"writer,omitempty"`         // Writers
	Mixer          []string `json:"mixer,omitempty"`          // Mixers
}

// Rights represents the streaming rights for the song.
type Rights struct {
	StreamAdsAvailable *bool   `json:"STREAM_ADS_AVAILABLE,omitempty"` // Ads available for streaming
	StreamAds          *string `json:"STREAM_ADS,omitempty"`           // Ads details
	StreamSubAvailable *bool   `json:"STREAM_SUB_AVAILABLE,omitempty"` // Subscription available for streaming
	StreamSub          *string `json:"STREAM_SUB,omitempty"`           // Subscription details
}

// SongType represents detailed information about a song.
type SongType struct {
	ALB_ID                 string               `json:"ALB_ID"`                     // '302127'
	ALB_TITLE              string               `json:"ALB_TITLE"`                  // 'Discovery'
	ALB_PICTURE            string               `json:"ALB_PICTURE"`                // '2e018122cb56986277102d2041a592c8'
	ARTISTS                []ArtistType         `json:"ARTISTS"`                    // List of artists
	ART_ID                 string               `json:"ART_ID"`                     // '27'
	ART_NAME               string               `json:"ART_NAME"`                   // 'Daft Punk'
	ARTIST_IS_DUMMY        bool                 `json:"ARTIST_IS_DUMMY"`            // false
	ART_PICTURE            string               `json:"ART_PICTURE"`                // 'f2bc007e9133c946ac3c3907ddc5d2ea'
	DATE_START             string               `json:"DATE_START"`                 // '0000-00-00'
	DISK_NUMBER            StringOrInt          `json:"DISK_NUMBER,omitempty"`      // '1'
	DURATION               StringOrInt          `json:"DURATION"`                   // '224'
	EXPLICIT_TRACK_CONTENT ExplicitTrackContent `json:"EXPLICIT_TRACK_CONTENT"`     // Explicit content status
	ISRC                   string               `json:"ISRC"`                       // 'GBDUW0000059'
	LYRICS_ID              int                  `json:"LYRICS_ID"`                  // 2780622
	LYRICS                 *LyricsType          `json:"LYRICS,omitempty"`           // Lyrics information
	EXPLICIT_LYRICS        *StringOrBool        `json:"EXPLICIT_LYRICS,omitempty"`  // Optional explicit lyrics status
	RANK                   string               `json:"RANK"`                       // '787708'
	SMARTRADIO             StringOrInt          `json:"SMARTRADIO"`                 // Can be '0' or 0
	SNG_ID                 string               `json:"SNG_ID"`                     // '3135556'
	SNG_TITLE              string               `json:"SNG_TITLE"`                  // 'Harder, Better, Faster, Stronger'
	SNG_CONTRIBUTORS       *SongContributors    `json:"SNG_CONTRIBUTORS,omitempty"` // Song contributors
	STATUS                 int                  `json:"STATUS"`                     // 3
	S_MOD                  int                  `json:"S_MOD"`                      // 0
	S_PREMIUM              int                  `json:"S_PREMIUM"`                  // 0
	TRACK_NUMBER           StringOrInt          `json:"TRACK_NUMBER"`               // '4'
	URL_REWRITING          string               `json:"URL_REWRITING"`              // 'daft-punk'
	VERSION                *string              `json:"VERSION,omitempty"`          // '(Extended Club Mix Edit)'
	MD5_ORIGIN             string               `json:"MD5_ORIGIN"`                 // '51afcde9f56a132096c0496cc95eb24b'
	FILESIZE_AAC_64        StringOrInt          `json:"FILESIZE_AAC_64"`            // Can be '0' or 0
	FILESIZE_MP3_64        StringOrInt          `json:"FILESIZE_MP3_64"`            // Can be '1798059' or 1798059
	FILESIZE_MP3_128       StringOrInt          `json:"FILESIZE_MP3_128"`           // Can be '3596119' or 3596119
	FILESIZE_MP3_256       StringOrInt          `json:"FILESIZE_MP3_256"`           // Can be '0' or 0
	FILESIZE_MP3_320       StringOrInt          `json:"FILESIZE_MP3_320"`           // Can be '0' or 0
	FILESIZE_MP4_RA1       StringOrInt          `json:"FILESIZE_MP4_RA1"`           // Can be '0' or 0
	FILESIZE_MP4_RA2       StringOrInt          `json:"FILESIZE_MP4_RA2"`           // Can be '0' or 0
	FILESIZE_MP4_RA3       StringOrInt          `json:"FILESIZE_MP4_RA3"`           // Can be '0' or 0
	FILESIZE_FLAC          StringOrInt          `json:"FILESIZE_FLAC"`              // Can be '0' or 0
	FILESIZE               StringOrInt          `json:"FILESIZE"`                   // Can be '3596119' or 3596119
	GAIN                   string               `json:"GAIN"`                       // '-12.4'
	MEDIA_VERSION          string               `json:"MEDIA_VERSION"`              // '8'
	TRACK_TOKEN            string               `json:"TRACK_TOKEN"`                // 'Track Token'
	TRACK_TOKEN_EXPIRE     int                  `json:"TRACK_TOKEN_EXPIRE"`         // 1614065380
	MEDIA                  []MediaType          `json:"MEDIA"`                      // Array of media details
	RIGHTS                 Rights               `json:"RIGHTS"`                     // Rights details
	PROVIDER_ID            string               `json:"PROVIDER_ID"`                // '3'
	Type                   string               `json:"__TYPE__"`                   // 'song'
}

// TrackType represents detailed information about a track including song and fallback data.
type TrackType struct {
	SongType
	FALLBACK       *SongType `json:"FALLBACK,omitempty"`       // Fallback song type
	TRACK_POSITION *int      `json:"TRACK_POSITION,omitempty"` // Track position
}

// ContributorsPublicAPI represents information about contributors from the public API.
type ContributorsPublicAPI struct {
	ID            int    `json:"id"`             // 27
	Name          string `json:"name"`           // 'Daft Punk'
	Link          string `json:"link"`           // 'https://www.deezer.com/artist/27'
	Share         string `json:"share"`          // 'https://www.deezer.com/artist/27?utm_source=deezer&utm_content=artist-27&utm_term=0_1614937516&utm_medium=web'
	Picture       string `json:"picture"`        // 'https://api.deezer.com/artist/27/image'
	PictureSmall  string `json:"picture_small"`  // 'https://e-cdns-images.dzcdn.net/images/artist/f2bc007e9133c946ac3c3907ddc5d2ea/56x56-000000-80-0-0.jpg'
	PictureMedium string `json:"picture_medium"` // 'https://e-cdns-images.dzcdn.net/images/artist/f2bc007e9133c946ac3c3907ddc5d2ea/250x250-000000-80-0-0.jpg'
	PictureBig    string `json:"picture_big"`    // 'https://e-cdns-images.dzcdn.net/images/artist/f2bc007e9133c946ac3c3907ddc5d2ea/500x500-000000-80-0-0.jpg'
	PictureXL     string `json:"picture_xl"`     // 'https://e-cdns-images.dzcdn.net/images/artist/f2bc007e9133c946ac3c3907ddc5d2ea/1000x1000-000000-80-0-0.jpg'
	Radio         bool   `json:"radio"`          // Radio availability
	Tracklist     string `json:"tracklist"`      // 'https://api.deezer.com/artist/27/top?limit=50'
	Type          string `json:"type"`           // 'artist'
	Role          string `json:"role"`           // 'Main'
}

// TrackTypePublicAPI represents public API data for a track including details about the album and artist.
type TrackTypePublicAPI struct {
	ID                    int                     `json:"id"`                      // 3135556
	Readable              bool                    `json:"readable"`                // Readable status
	Title                 string                  `json:"title"`                   // 'Harder, Better, Faster, Stronger'
	TitleShort            string                  `json:"title_short"`             // 'Harder, Better, Faster, Stronger'
	TitleVersion          *string                 `json:"title_version,omitempty"` // Optional version title
	ISRC                  string                  `json:"isrc"`                    // 'GBDUW0000059'
	Link                  string                  `json:"link"`                    // 'https://www.deezer.com/track/3135556'
	Share                 string                  `json:"share"`                   // Share link
	Duration              StringOrInt             `json:"duration"`                // 224
	TrackPosition         int                     `json:"track_position"`          // 4
	DiskNumber            StringOrInt             `json:"disk_number"`             // 1
	Rank                  int                     `json:"rank"`                    // 956167
	ReleaseDate           string                  `json:"release_date"`            // '2001-03-07'
	ExplicitLyrics        *StringOrBool           `json:"explicit_lyrics"`         // Explicit lyrics status
	ExplicitContentLyrics int                     `json:"explicit_content_lyrics"` // Explicit content lyrics status
	ExplicitContentCover  int                     `json:"explicit_content_cover"`  // Explicit content cover status
	Preview               string                  `json:"preview"`                 // Preview link
	BPM                   float64                 `json:"bpm"`                     // 123.4
	Gain                  float64                 `json:"gain"`                    // -12.4
	AvailableCountries    []string                `json:"available_countries"`     // List of available countries
	Contributors          []ContributorsPublicAPI `json:"contributors"`            // List of contributors
	MD5Image              string                  `json:"md5_image"`               // MD5 image hash
	Artist                ContributorsPublicAPI   `json:"artist"`                  // Artist details
	Album                 struct {
		ID          int    `json:"id"`           // 302127
		Title       string `json:"title"`        // 'Discovery'
		Link        string `json:"link"`         // 'https://www.deezer.com/album/302127'
		Cover       string `json:"cover"`        // 'https://api.deezer.com/album/302127/image'
		CoverSmall  string `json:"cover_small"`  // 'https://e-cdns-images.dzcdn.net/images/cover/2e018122cb56986277102d2041a592c8/56x56-000000-80-0-0.jpg'
		CoverMedium string `json:"cover_medium"` // 'https://e-cdns-images.dzcdn.net/images/cover/2e018122cb56986277102d2041a592c8/250x250-000000-80-0-0.jpg'
		CoverBig    string `json:"cover_big"`    // 'https://e-cdns-images.dzcdn.net/images/cover/2e018122cb56986277102d2041a592c8/500x500-000000-80-0-0.jpg'
		CoverXL     string `json:"cover_xl"`     // 'https://e-cdns-images.dzcdn.net/images/cover/2e018122cb56986277102d2041a592c8/1000x1000-000000-80-0-0.jpg'
		MD5Image    string `json:"md5_image"`    // '2e018122cb56986277102d2041a592c8'
		ReleaseDate string `json:"release_date"` // '2001-03-07'
		Tracklist   string `json:"tracklist"`    // 'https://api.deezer.com/album/302127/tracks'
		Type        string `json:"type"`         // 'album'
	} `json:"album"`
	Type string `json:"type"` // 'track'
}

// UnmarshalJSON for SongContributors allows dynamic handling of multiple structures.
func (sc *SongContributors) UnmarshalJSON(data []byte) error {
	// Attempt to unmarshal directly into the struct
	type Alias SongContributors
	var tmp Alias
	if err := json.Unmarshal(data, &tmp); err == nil {
		*sc = SongContributors(tmp)
		return nil
	}

	// If the above fails, try parsing as an empty array
	if string(data) == "[]" || string(data) == "{}" {
		*sc = SongContributors{}
		return nil
	}

	return fmt.Errorf("failed to unmarshal SongContributors: %s", string(data))
}
