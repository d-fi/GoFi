package types

type PlaylistInfoMinimal struct {
	PlaylistID        string `json:"PLAYLIST_ID"`
	ParentPlaylistID  string `json:"PARENT_PLAYLIST_ID"`
	Type              string `json:"TYPE"`
	Title             string `json:"TITLE"`
	ParentUserID      string `json:"PARENT_USER_ID"`
	ParentUsername    string `json:"PARENT_USERNAME"`
	ParentUserPicture string `json:"PARENT_USER_PICTURE,omitempty"`
	Status            string `json:"STATUS"`
	PlaylistPicture   string `json:"PLAYLIST_PICTURE"`
	PictureType       string `json:"PICTURE_TYPE"`
	NbSong            int    `json:"NB_SONG"`
	HasArtistLinked   bool   `json:"HAS_ARTIST_LINKED"`
	DateAdd           string `json:"DATE_ADD"`
	DateMod           string `json:"DATE_MOD"`
	TYPE_INTERNAL     string `json:"__TYPE__"`
}

type PlaylistInfo struct {
	PlaylistID        string `json:"PLAYLIST_ID"`
	Description       string `json:"DESCRIPTION,omitempty"`
	ParentUsername    string `json:"PARENT_USERNAME"`
	ParentUserPicture string `json:"PARENT_USER_PICTURE,omitempty"`
	ParentUserID      string `json:"PARENT_USER_ID"`
	PictureType       string `json:"PICTURE_TYPE"`
	PlaylistPicture   string `json:"PLAYLIST_PICTURE"`
	Title             string `json:"TITLE"`
	Type              string `json:"TYPE"`
	Status            string `json:"STATUS"`
	UserID            string `json:"USER_ID"`
	DateAdd           string `json:"DATE_ADD"`
	DateMod           string `json:"DATE_MOD"`
	DateCreate        string `json:"DATE_CREATE"`
	NbSong            int    `json:"NB_SONG"`
	NbFan             int    `json:"NB_FAN"`
	Checksum          string `json:"CHECKSUM"`
	HasArtistLinked   bool   `json:"HAS_ARTIST_LINKED"`
	IsSponsored       bool   `json:"IS_SPONSORED"`
	IsEdito           bool   `json:"IS_EDITO"`
	TYPE_INTERNAL     string `json:"__TYPE__"`
}

type PlaylistTracksType struct {
	Data          []TrackType `json:"data"`
	Count         int         `json:"count"`
	Total         int         `json:"total"`
	FilteredCount int         `json:"filtered_count"`
	FilteredItems []int       `json:"filtered_items,omitempty"`
	Next          *int        `json:"next,omitempty"`
}
