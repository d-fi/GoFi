package types

// GeneresType represents genre details, including ID and name.
type GeneresType struct {
	GenreID   int    `json:"GENRE_ID"`   // Genre ID, e.g., 232
	GenreName string `json:"GENRE_NAME"` // Genre name, e.g., 'Technology'
}

// ShowEpisodeType represents information about a show episode.
type ShowEpisodeType struct {
	EpisodeID                string `json:"EPISODE_ID"`                  // Episode ID, e.g., '294961882'
	EpisodeStatus            string `json:"EPISODE_STATUS"`              // Episode status, e.g., '1'
	Available                bool   `json:"AVAILABLE"`                   // Availability status
	ShowID                   string `json:"SHOW_ID"`                     // Show ID, e.g., '1235862'
	ShowName                 string `json:"SHOW_NAME"`                   // Show name, e.g., 'Masters of Scale with Reid Hoffman'
	ShowArtMD5               string `json:"SHOW_ART_MD5"`                // Show art MD5 hash, e.g., '52d6e09bccf1369d5758e7a45ee98b7e'
	ShowDescription          string `json:"SHOW_DESCRIPTION"`            // Show description, detailed text about the show
	ShowIsExplicit           string `json:"SHOW_IS_EXPLICIT"`            // Explicit content flag, e.g., '2'
	EpisodeTitle             string `json:"EPISODE_TITLE"`               // Episode title, e.g., '87. Frustration is your friend, w/Houzz founder Adi Tatarko'
	EpisodeDescription       string `json:"EPISODE_DESCRIPTION"`         // Episode description, detailed text about the episode
	MD5Origin                string `json:"MD5_ORIGIN"`                  // MD5 origin hash, e.g., ''
	FilesizeMP332            string `json:"FILESIZE_MP3_32"`             // MP3 file size for 32 kbps, e.g., '0'
	FilesizeMP364            string `json:"FILESIZE_MP3_64"`             // MP3 file size for 64 kbps, e.g., '0'
	EpisodeDirectStreamURL   string `json:"EPISODE_DIRECT_STREAM_URL"`   // Direct stream URL for the episode
	ShowIsDirectStream       string `json:"SHOW_IS_DIRECT_STREAM"`       // Direct stream flag, e.g., '1'
	Duration                 string `json:"DURATION"`                    // Duration of the episode, e.g., '2022'
	EpisodePublishedTime     string `json:"EPISODE_PUBLISHED_TIMESTAMP"` // Published timestamp, e.g., '2021-04-20 09:00:00'
	EpisodeUpdateTime        string `json:"EPISODE_UPDATE_TIMESTAMP"`    // Update timestamp, e.g., '2021-04-20 10:23:21'
	ShowIsAdvertisingAllowed string `json:"SHOW_IS_ADVERTISING_ALLOWED"` // Advertising allowed flag, e.g., '1'
	ShowIsDownloadAllowed    string `json:"SHOW_IS_DOWNLOAD_ALLOWED"`    // Download allowed flag, e.g., '1'
	TrackToken               string `json:"TRACK_TOKEN"`                 // Track token for the episode
	TrackTokenExpire         int    `json:"TRACK_TOKEN_EXPIRE"`          // Track token expiration timestamp, e.g., 1619011217
	Type                     string `json:"__TYPE__"`                    // Type of content, e.g., 'episode'
}

// ShowType represents detailed information about a show, including episodes and metadata.
type ShowType struct {
	Data struct {
		Available                bool          `json:"AVAILABLE"`                   // Availability status
		ShowIsExplicit           string        `json:"SHOW_IS_EXPLICIT"`            // Explicit content flag, e.g., '2'
		LabelID                  string        `json:"LABEL_ID"`                    // Label ID, e.g., '35611'
		LabelName                string        `json:"LABEL_NAME"`                  // Label name, e.g., 'Art19'
		LanguageCD               string        `json:"LANGUAGE_CD"`                 // Language code, e.g., 'en'
		ShowIsDirectStream       string        `json:"SHOW_IS_DIRECT_STREAM"`       // Direct stream flag, e.g., '1'
		ShowIsAdvertisingAllowed string        `json:"SHOW_IS_ADVERTISING_ALLOWED"` // Advertising allowed flag, e.g., '1'
		ShowIsDownloadAllowed    string        `json:"SHOW_IS_DOWNLOAD_ALLOWED"`    // Download allowed flag, e.g., '1'
		ShowEpisodeDisplayCount  string        `json:"SHOW_EPISODE_DISPLAY_COUNT"`  // Number of displayed episodes, e.g., '0'
		ShowID                   string        `json:"SHOW_ID"`                     // Show ID, e.g., '1235862'
		ShowArtMD5               string        `json:"SHOW_ART_MD5"`                // Show art MD5 hash, e.g., '52d6e09bccf1369d5758e7a45ee98b7e'
		ShowName                 string        `json:"SHOW_NAME"`                   // Show name, e.g., 'Masters of Scale with Reid Hoffman'
		ShowDescription          string        `json:"SHOW_DESCRIPTION"`            // Show description, detailed text about the show
		ShowStatus               string        `json:"SHOW_STATUS"`                 // Show status, e.g., '1'
		ShowType                 string        `json:"SHOW_TYPE"`                   // Show type, e.g., '0'
		Genres                   []GeneresType `json:"GENRES"`                      // List of genres associated with the show
		NBFan                    int           `json:"NB_FAN"`                      // Number of fans, e.g., 658
		NBRate                   int           `json:"NB_RATE"`                     // Number of ratings, e.g., 0
		Rating                   string        `json:"RATING"`                      // Show rating, e.g., '0'
		Type                     string        `json:"__TYPE__"`                    // Type of content, e.g., 'show'
	} `json:"DATA"`
	FavoriteStatus bool `json:"FAVORITE_STATUS"` // Favorite status flag, e.g., false
	Episodes       struct {
		Data          []ShowEpisodeType `json:"data"`           // List of episodes
		Count         int               `json:"count"`          // Number of episodes in the list, e.g., 1
		Total         int               `json:"total"`          // Total number of episodes available, e.g., 174
		FilteredCount int               `json:"filtered_count"` // Number of filtered episodes, e.g., 0
	} `json:"EPISODES"`
}
