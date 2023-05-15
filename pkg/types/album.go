package types

type AlbumTypeMinimal struct {
	ALB_ID                 string       `json:"ALB_ID"`
	ALB_TITLE              string       `json:"ALB_TITLE"`
	ALB_PICTURE            string       `json:"ALB_PICTURE"`
	ARTISTS                []ArtistType `json:"ARTISTS"`
	AVAILABLE              bool         `json:"AVAILABLE"`
	VERSION                string       `json:"VERSION"`
	ART_ID                 string       `json:"ART_ID"`
	ART_NAME               string       `json:"ART_NAME"`
	EXPLICIT_ALBUM_CONTENT struct {
		EXPLICIT_LYRICS_STATUS int `json:"EXPLICIT_LYRICS_STATUS"`
		EXPLICIT_COVER_STATUS  int `json:"EXPLICIT_COVER_STATUS"`
	} `json:"EXPLICIT_ALBUM_CONTENT"`
	PHYSICAL_RELEASE_DATE string `json:"PHYSICAL_RELEASE_DATE"`
	TYPE                  string `json:"TYPE"`
	ARTIST_IS_DUMMY       bool   `json:"ARTIST_IS_DUMMY"`
	NUMBER_TRACK          int    `json:"NUMBER_TRACK"`
	__TYPE__              string `json:"__TYPE__"`
}

type AlbumType struct {
	ALB_CONTRIBUTORS struct {
		MainArtist []string `json:"main_artist"`
	} `json:"ALB_CONTRIBUTORS"`
	ALB_ID                 string `json:"ALB_ID"`
	ALB_PICTURE            string `json:"ALB_PICTURE"`
	EXPLICIT_ALBUM_CONTENT struct {
		EXPLICIT_LYRICS_STATUS int `json:"EXPLICIT_LYRICS_STATUS"`
		EXPLICIT_COVER_STATUS  int `json:"EXPLICIT_COVER_STATUS"`
	} `json:"EXPLICIT_ALBUM_CONTENT"`
	ALB_TITLE             string       `json:"ALB_TITLE"`
	ARTISTS               []ArtistType `json:"ARTISTS"`
	ART_ID                string       `json:"ART_ID"`
	ART_NAME              string       `json:"ART_NAME"`
	ARTIST_IS_DUMMY       bool         `json:"ARTIST_IS_DUMMY"`
	DIGITAL_RELEASE_DATE  string       `json:"DIGITAL_RELEASE_DATE"`
	EXPLICIT_LYRICS       *string      `json:"EXPLICIT_LYRICS,omitempty"`
	NB_FAN                int          `json:"NB_FAN"`
	NUMBER_DISK           string       `json:"NUMBER_DISK"`
	NUMBER_TRACK          string       `json:"NUMBER_TRACK"`
	PHYSICAL_RELEASE_DATE *string      `json:"PHYSICAL_RELEASE_DATE,omitempty"`
	PRODUCER_LINE         string       `json:"PRODUCER_LINE"`
	PROVIDER_ID           string       `json:"PROVIDER_ID"`
	RANK                  string       `json:"RANK"`
	RANK_ART              string       `json:"RANK_ART"`
	STATUS                string       `json:"STATUS"`
	TYPE                  string       `json:"TYPE"`
	UPC                   string       `json:"UPC"`
	__TYPE__              string       `json:"__TYPE__"`
}

type AlbumTracksType struct {
	Data          []TrackType `json:"data"`
	Count         int         `json:"count"`
	Total         int         `json:"total"`
	FilteredCount int         `json:"filtered_count"`
	FilteredItems []int       `json:"filtered_items,omitempty"`
	Next          int         `json:"next,omitempty"`
}

type TrackDataPublicApi struct {
	ID                    int              `json:"id"`
	Readable              bool             `json:"readable"`
	Title                 string           `json:"title"`
	TitleShort            string           `json:"title_short"`
	TitleVersion          string           `json:"title_version,omitempty"`
	Link                  string           `json:"link"`
	Duration              int              `json:"duration"`
	Rank                  int              `json:"rank"`
	ExplicitLyrics        bool             `json:"explicit_lyrics"`
	ExplicitContentLyrics int              `json:"explicit_content_lyrics"`
	ExplicitContentCover  int              `json:"explicit_content_cover"`
	Preview               string           `json:"preview"`
	MD5Image              string           `json:"md5_image"`
	Artist                ArtistDataPublic `json:"artist"`
	Type                  string           `json:"type"`
}

type ArtistDataPublic struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Tracklist string `json:"tracklist"`
	Type      string `json:"type"`
}

type GenreTypePublicApi struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
	Type    string `json:"type"`
}

type AlbumTypePublicApi struct {
	ID                    int                     `json:"id"`
	Title                 string                  `json:"title"`
	UPC                   string                  `json:"upc"`
	Link                  string                  `json:"link"`
	Share                 string                  `json:"share"`
	Cover                 string                  `json:"cover"`
	CoverSmall            string                  `json:"cover_small"`
	CoverMedium           string                  `json:"cover_medium"`
	CoverBig              string                  `json:"cover_big"`
	CoverXL               string                  `json:"cover_xl"`
	MD5Image              string                  `json:"md5_image"`
	GenreID               int                     `json:"genre_id"`
	Genres                GenreTypePublicApiList  `json:"genres"`
	Label                 string                  `json:"label"`
	NbTracks              int                     `json:"nb_tracks"`
	Duration              int                     `json:"duration"`
	Fans                  int                     `json:"fans"`
	Rating                int                     `json:"rating"`
	ReleaseDate           string                  `json:"release_date"`
	RecordType            string                  `json:"record_type"`
	Available             bool                    `json:"available"`
	Tracklist             string                  `json:"tracklist"`
	ExplicitLyrics        bool                    `json:"explicit_lyrics"`
	ExplicitContentLyrics int                     `json:"explicit_content_lyrics"`
	ExplicitContentCover  int                     `json:"explicit_content_cover"`
	Contributors          []ContributorsPublicApi `json:"contributors"`
	Artist                ContributorsPublicApi   `json:"artist"`
	Type                  string                  `json:"type"`
	Tracks                TrackDataPublicApiList  `json:"tracks"`
}

type GenreTypePublicApiList struct {
	Data []GenreTypePublicApi `json:"data"`
}

type ContributorsPublicApi struct {
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
}

type TrackDataPublicApiList struct {
	Data []TrackDataPublicApi `json:"data"`
}
