package types

type SearchTypeCommon struct {
	Count         int   `json:"count"`
	Total         int   `json:"total"`
	FilteredCount int   `json:"filtered_count"`
	FilteredItems []int `json:"filtered_items"`
	Next          int   `json:"next"`
}

type AlbumSearchType struct {
	SearchTypeCommon
	Data []AlbumTypeMinimal `json:"data"`
}

type ArtistSearchType struct {
	SearchTypeCommon
	Data []ArtistInfoTypeMinimal `json:"data"`
}

type PlaylistSearchType struct {
	SearchTypeCommon
	Data []PlaylistInfoMinimal `json:"data"`
}

type TrackSearchType struct {
	SearchTypeCommon
	Data []TrackType `json:"data"`
}

type ProfileSearchType struct {
	SearchTypeCommon
	Data []ProfileTypeMinimal `json:"data"`
}

type RadioSearchType struct {
	SearchTypeCommon
	Data []RadioType `json:"data"`
}

type LiveSearchType struct {
	SearchTypeCommon
	Data []interface{} `json:"data"`
}

type ShowSearchType struct {
	SearchTypeCommon
	Data []ShowEpisodeType `json:"data"`
}

type DiscographyType struct {
	Data          []AlbumType `json:"data"`
	Count         int         `json:"count"`
	Total         int         `json:"total"`
	CacheVersion  int         `json:"cache_version"`
	FilteredCount int         `json:"filtered_count"`
	ArtID         int         `json:"art_id"`
	Start         int         `json:"start"`
	NB            int         `json:"nb"`
}

type SearchType struct {
	QUERY       string             `json:"QUERY"`
	FUZZINNESS  bool               `json:"FUZZINNESS"`
	AUTOCORRECT bool               `json:"AUTOCORRECT"`
	TOP_RESULT  []interface{}      `json:"TOP_RESULT"`
	ORDER       []string           `json:"ORDER"`
	ALBUM       AlbumSearchType    `json:"ALBUM"`
	ARTIST      ArtistSearchType   `json:"ARTIST"`
	TRACK       TrackSearchType    `json:"TRACK"`
	PLAYLIST    PlaylistSearchType `json:"PLAYLIST"`
	RADIO       RadioSearchType    `json:"RADIO"`
	SHOW        ShowSearchType     `json:"SHOW"`
	USER        ProfileSearchType  `json:"USER"`
	LIVESTREAM  LiveSearchType     `json:"LIVESTREAM"`
	CHANNEL     ChannelSearchType  `json:"CHANNEL"`
}
