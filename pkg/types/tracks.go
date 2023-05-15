package types

type MediaType struct {
	TYPE string `json:"TYPE"`
	HREF string `json:"HREF"`
}

type LyricsSync struct {
	LrcTimestamp string `json:"lrc_timestamp"`
	Milliseconds string `json:"milliseconds"`
	Duration     string `json:"duration"`
	Line         string `json:"line"`
}

type LyricsType struct {
	LYRICS_ID         *string      `json:"LYRICS_ID,omitempty"`
	LYRICS_SYNC_JSON  []LyricsSync `json:"LYRICS_SYNC_JSON,omitempty"`
	LYRICS_TEXT       string       `json:"LYRICS_TEXT"`
	LYRICS_COPYRIGHTS *string      `json:"LYRICS_COPYRIGHTS,omitempty"`
	LYRICS_WRITERS    *string      `json:"LYRICS_WRITERS,omitempty"`
}

type ExplicitTrackContent struct {
	ExplicitLyricsStatus int `json:"EXPLICIT_LYRICS_STATUS"`
	ExplicitCoverStatus  int `json:"EXPLICIT_COVER_STATUS"`
}

type SongContributors struct {
	MainArtist     []string `json:"main_artist,omitempty"`
	Author         []string `json:"author,omitempty"`
	Composer       []string `json:"composer,omitempty"`
	MusicPublisher []string `json:"musicpublisher,omitempty"`
	Producer       []string `json:"producer,omitempty"`
	Publisher      []string `json:"publisher"`
	Engineer       []string `json:"engineer,omitempty"`
	Writer         []string `json:"writer,omitempty"`
	Mixer          []string `json:"mixer,omitempty"`
}

type Rights struct {
	StreamAdsAvailable *bool   `json:"STREAM_ADS_AVAILABLE,omitempty"`
	StreamAds          *string `json:"STREAM_ADS,omitempty"`
	StreamSubAvailable *bool   `json:"STREAM_SUB_AVAILABLE,omitempty"`
	StreamSub          *string `json:"STREAM_SUB,omitempty"`
}

type SongType struct {
	ALB_ID                 string               `json:"ALB_ID"`
	ALB_TITLE              string               `json:"ALB_TITLE"`
	ALB_PICTURE            string               `json:"ALB_PICTURE"`
	ARTISTS                []ArtistType         `json:"ARTISTS"`
	ART_ID                 string               `json:"ART_ID"`
	ART_NAME               string               `json:"ART_NAME"`
	ARTIST_IS_DUMMY        bool                 `json:"ARTIST_IS_DUMMY"`
	ART_PICTURE            string               `json:"ART_PICTURE"`
	DATE_START             string               `json:"DATE_START"`
	DISK_NUMBER            *string              `json:"DISK_NUMBER,omitempty"`
	DURATION               string               `json:"DURATION"`
	EXPLICIT_TRACK_CONTENT ExplicitTrackContent `json:"EXPLICIT_TRACK_CONTENT"`
	ISRC                   string               `json:"ISRC"`
	LYRICS_ID              int                  `json:"LYRICS_ID"`
	LYRICS                 *LyricsType          `json:"LYRICS,omitempty"`
	EXPLICIT_LYRICS        *string              `json:"EXPLICIT_LYRICS,omitempty"`
	RANK                   string               `json:"RANK"`
	SMARTRADIO             string               `json:"SMARTRADIO"`
	SNG_ID                 string               `json:"SNG_ID"`
	SNG_TITLE              string               `json:"SNG_TITLE"`
	SNG_CONTRIBUTORS       SongContributors     `json:"SNG_CONTRIBUTORS,omitempty"`
	STATUS                 int                  `json:"STATUS"`
	S_MOD                  int                  `json:"S_MOD"`
	S_PREMIUM              int                  `json:"S_PREMIUM"`
	TRACK_NUMBER           int                  `json:"TRACK_NUMBER"`
	URL_REWRITING          string               `json:"URL_REWRITING"`
	VERSION                *string              `json:"VERSION,omitempty"`
	MD5_ORIGIN             string               `json:"MD5_ORIGIN"`
	FILESIZE_AAC_64        string               `json:"FILESIZE_AAC_64"`
	FILESIZE_MP3_64        string               `json:"FILESIZE_MP3_64"`
	FILESIZE_MP3_128       string               `json:"FILESIZE_MP3_128"`
	FILESIZE_MP3_256       string               `json:"FILESIZE_MP3_256"`
	FILESIZE_MP3_320       string               `json:"FILESIZE_MP3_320"`
	FILESIZE_MP4_RA1       string               `json:"FILESIZE_MP4_RA1"`
	FILESIZE_MP4_RA2       string               `json:"FILESIZE_MP4_RA2"`
	FILESIZE_MP4_RA3       string               `json:"FILESIZE_MP4_RA3"`
	FILESIZE_FLAC          string               `json:"FILESIZE_FLAC"`
	FILESIZE               string               `json:"FILESIZE"`
	GAIN                   string               `json:"GAIN"`
	MEDIA_VERSION          string               `json:"MEDIA_VERSION"`
	TRACK_TOKEN            string               `json:"TRACK_TOKEN"`
	TRACK_TOKEN_EXPIRE     int                  `json:"TRACK_TOKEN_EXPIRE"`
	MEDIA                  []MediaType          `json:"MEDIA"`
	RIGHTS                 Rights               `json:"RIGHTS"`
	PROVIDER_ID            string               `json:"PROVIDER_ID"`
	Type                   string               `json:"__TYPE__"`
}

type TrackType struct {
	SongType
	FALLBACK       *SongType `json:"FALLBACK,omitempty"`
	TRACK_POSITION *int      `json:"TRACK_POSITION,omitempty"`
}

type ContributorsPublicAPI struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Link          string `json:"link"`
	Share         string `json:"share"`
	Picture       string `json:"picture"`
	PictureSmall  string `json:"picture_small"`
	PictureMedium string `json:"picture_medium"`
	PictureBig    string `json:"picture_big"`
	PictureXL     string `json:"picture_xl"`
	Radio         bool   `json:"radio"`
	Tracklist     string `json:"tracklist"`
	Type          string `json:"type"`
	Role          string `json:"role"`
}

type TrackTypePublicAPI struct {
	ID                    int                     `json:"id"`
	Readable              bool                    `json:"readable"`
	Title                 string                  `json:"title"`
	TitleShort            string                  `json:"title_short"`
	TitleVersion          *string                 `json:"title_version,omitempty"`
	ISRC                  string                  `json:"isrc"`
	Link                  string                  `json:"link"`
	Share                 string                  `json:"share"`
	Duration              int                     `json:"duration"`
	TrackPosition         int                     `json:"track_position"`
	DiskNumber            int                     `json:"disk_number"`
	Rank                  int                     `json:"rank"`
	ReleaseDate           string                  `json:"release_date"`
	ExplicitLyrics        bool                    `json:"explicit_lyrics"`
	ExplicitContentLyrics int                     `json:"explicit_content_lyrics"`
	ExplicitContentCover  int                     `json:"explicit_content_cover"`
	Preview               string                  `json:"preview"`
	BPM                   float64                 `json:"bpm"`
	Gain                  float64                 `json:"gain"`
	AvailableCountries    []string                `json:"available_countries"`
	Contributors          []ContributorsPublicAPI `json:"contributors"`
	MD5Image              string                  `json:"md5_image"`
	Artist                ContributorsPublicAPI   `json:"artist"`
	Album                 struct {
		ID          int    `json:"id"`
		Title       string `json:"title"`
		Link        string `json:"link"`
		Cover       string `json:"cover"`
		CoverSmall  string `json:"cover_small"`
		CoverMedium string `json:"cover_medium"`
		CoverBig    string `json:"cover_big"`
		CoverXL     string `json:"cover_xl"`
		MD5Image    string `json:"md5_image"`
		ReleaseDate string `json:"release_date"`
		Tracklist   string `json:"tracklist"`
		Type        string `json:"type"`
	} `json:"album"`
	Type string `json:"type"`
}
