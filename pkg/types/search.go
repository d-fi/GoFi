package types

type SearchTypeCommon struct {
	count          int   `json:"count"`
	total          int   `json:"total"`
	filtered_count int   `json:"filtered_count"`
	filtered_items []int `json:"filtered_items"`
	next           int   `json:"next"`
}

type AlbumSearchType struct {
	SearchTypeCommon
	data []AlbumTypeMinimal `json:"data"`
}

type ArtistSearchType struct {
	SearchTypeCommon
	data []ArtistInfoTypeMinimal `json:"data"`
}

type PlaylistSearchType struct {
	SearchTypeCommon
	data []PlaylistInfoMinimal `json:"data"`
}

type TrackSearchType struct {
	SearchTypeCommon
	data []TrackType `json:"data"`
}

type ProfileSearchType struct {
	SearchTypeCommon
	data []ProfileTypeMinimal `json:"data"`
}

type RadioSearchType struct {
	SearchTypeCommon
	data []RadioType `json:"data"`
}

type LiveSearchType struct {
	SearchTypeCommon
	data []interface{} `json:"data"`
}

type ShowSearchType struct {
	SearchTypeCommon
	data []ShowEpisodeType `json:"data"`
}

type DiscographyType struct {
	data           []AlbumType `json:"data"`
	count          int         `json:"count"`
	total          int         `json:"total"`
	cache_version  int         `json:"cache_version"`
	filtered_count int         `json:"filtered_count"`
	art_id         int         `json:"art_id"`
	start          int         `json:"start"`
	nb             int         `json:"nb"`
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
