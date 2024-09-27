package types

// PlaylistInfoMinimal represents basic information about a playlist.
type PlaylistInfoMinimal struct {
	PlaylistID        string      `json:"PLAYLIST_ID"`                   // '4523119944'
	ParentPlaylistID  string      `json:"PARENT_PLAYLIST_ID"`            // '0'
	Type              string      `json:"TYPE"`                          // '0'
	Title             string      `json:"TITLE"`                         // 'wtf playlist'
	ParentUserID      string      `json:"PARENT_USER_ID"`                // '2064440442'
	ParentUsername    string      `json:"PARENT_USERNAME"`               // 'sayem314'
	ParentUserPicture string      `json:"PARENT_USER_PICTURE,omitempty"` // Optional parent user picture
	Status            StringOrInt `json:"STATUS"`                        // '0'
	PlaylistPicture   string      `json:"PLAYLIST_PICTURE"`              // 'e206dafb59a3d378d7ffacc989bc4e35'
	PictureType       string      `json:"PICTURE_TYPE"`                  // 'playlist'
	NbSong            int         `json:"NB_SONG"`                       // 180
	HasArtistLinked   bool        `json:"HAS_ARTIST_LINKED"`             // True if artists are linked
	DateAdd           string      `json:"DATE_ADD"`                      // '2021-01-29 20:54:13'
	DateMod           string      `json:"DATE_MOD"`                      // '2021-02-01 05:52:40'
	TYPE_INTERNAL     string      `json:"__TYPE__"`                      // 'playlist'
}

// PlaylistInfo represents detailed information about a playlist.
type PlaylistInfo struct {
	PlaylistID        string      `json:"PLAYLIST_ID"`                   // '4523119944'
	Description       string      `json:"DESCRIPTION,omitempty"`         // Optional description
	ParentUsername    string      `json:"PARENT_USERNAME"`               // 'sayem314'
	ParentUserPicture string      `json:"PARENT_USER_PICTURE,omitempty"` // Optional parent user picture
	ParentUserID      string      `json:"PARENT_USER_ID"`                // '2064440442'
	PictureType       string      `json:"PICTURE_TYPE"`                  // 'cover'
	PlaylistPicture   string      `json:"PLAYLIST_PICTURE"`              // 'e206dafb59a3d378d7ffacc989bc4e35'
	Title             string      `json:"TITLE"`                         // 'wtf playlist'
	Type              string      `json:"TYPE"`                          // '0'
	Status            StringOrInt `json:"STATUS"`                        // '0'
	UserID            string      `json:"USER_ID"`                       // '2064440442'
	DateAdd           string      `json:"DATE_ADD"`                      // '2018-09-08 19:13:57'
	DateMod           string      `json:"DATE_MOD"`                      // '2018-09-08 19:14:11'
	DateCreate        string      `json:"DATE_CREATE"`                   // '2018-05-31 00:01:05'
	NbSong            int         `json:"NB_SONG"`                       // 3
	NbFan             int         `json:"NB_FAN"`                        // 0
	Checksum          string      `json:"CHECKSUM"`                      // 'c185d123834444e3c8869e235dd6f0a6'
	HasArtistLinked   bool        `json:"HAS_ARTIST_LINKED"`             // True if artists are linked
	IsSponsored       bool        `json:"IS_SPONSORED"`                  // True if the playlist is sponsored
	IsEdito           bool        `json:"IS_EDITO"`                      // True if the playlist is editorial
	TYPE_INTERNAL     string      `json:"__TYPE__"`                      // 'playlist'
}

// PlaylistTracksType represents track information within a playlist.
type PlaylistTracksType struct {
	Data          []TrackType `json:"data"`                     // Array of tracks
	Count         int         `json:"count"`                    // Number of tracks
	Total         int         `json:"total"`                    // Total number of tracks
	FilteredCount int         `json:"filtered_count"`           // Filtered track count
	FilteredItems []int       `json:"filtered_items,omitempty"` // Optional filtered track IDs
	Next          *int        `json:"next,omitempty"`           // Optional next track index
}
