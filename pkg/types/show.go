package types

type GeneresType struct {
	GenreID   int    `json:"GENRE_ID"`
	GenreName string `json:"GENRE_NAME"`
}

type ShowEpisodeType struct {
	EpisodeID                string `json:"EPISODE_ID"`
	EpisodeStatus            string `json:"EPISODE_STATUS"`
	Available                bool   `json:"AVAILABLE"`
	ShowID                   string `json:"SHOW_ID"`
	ShowName                 string `json:"SHOW_NAME"`
	ShowArtMD5               string `json:"SHOW_ART_MD5"`
	ShowDescription          string `json:"SHOW_DESCRIPTION"`
	ShowIsExplicit           string `json:"SHOW_IS_EXPLICIT"`
	EpisodeTitle             string `json:"EPISODE_TITLE"`
	EpisodeDescription       string `json:"EPISODE_DESCRIPTION"`
	MD5Origin                string `json:"MD5_ORIGIN"`
	FilesizeMP332            string `json:"FILESIZE_MP3_32"`
	FilesizeMP364            string `json:"FILESIZE_MP3_64"`
	EpisodeDirectStreamURL   string `json:"EPISODE_DIRECT_STREAM_URL"`
	ShowIsDirectStream       string `json:"SHOW_IS_DIRECT_STREAM"`
	Duration                 string `json:"DURATION"`
	EpisodePublishedTime     string `json:"EPISODE_PUBLISHED_TIMESTAMP"`
	EpisodeUpdateTime        string `json:"EPISODE_UPDATE_TIMESTAMP"`
	ShowIsAdvertisingAllowed string `json:"SHOW_IS_ADVERTISING_ALLOWED"`
	ShowIsDownloadAllowed    string `json:"SHOW_IS_DOWNLOAD_ALLOWED"`
	TrackToken               string `json:"TRACK_TOKEN"`
	TrackTokenExpire         string `json:"TRACK_TOKEN_EXPIRE"`
	Type                     string `json:"__TYPE__"`
}

type ShowType struct {
	Data struct {
		Available                bool          `json:"AVAILABLE"`
		ShowIsExplicit           string        `json:"SHOW_IS_EXPLICIT"`
		LabelID                  string        `json:"LABEL_ID"`
		LabelName                string        `json:"LABEL_NAME"`
		LanguageCD               string        `json:"LANGUAGE_CD"`
		ShowIsDirectStream       string        `json:"SHOW_IS_DIRECT_STREAM"`
		ShowIsAdvertisingAllowed string        `json:"SHOW_IS_ADVERTISING_ALLOWED"`
		ShowIsDownloadAllowed    string        `json:"SHOW_IS_DOWNLOAD_ALLOWED"`
		ShowEpisodeDisplayCount  string        `json:"SHOW_EPISODE_DISPLAY_COUNT"`
		ShowID                   string        `json:"SHOW_ID"`
		ShowArtMD5               string        `json:"SHOW_ART_MD5"`
		ShowName                 string        `json:"SHOW_NAME"`
		ShowDescription          string        `json:"SHOW_DESCRIPTION"`
		ShowStatus               string        `json:"SHOW_STATUS"`
		ShowType                 string        `json:"SHOW_TYPE"`
		Genres                   []GeneresType `json:"GENRES"`
		NBFan                    int           `json:"NB_FAN"`
		NBRate                   int           `json:"NB_RATE"`
		Rating                   string        `json:"RATING"`
		Type                     string        `json:"__TYPE__"`
	} `json:"DATA"`
	FavoriteStatus bool `json:"FAVORITE_STATUS"`
	Episodes       struct {
		Data          []ShowEpisodeType `json:"data"`
		Count         int               `json:"count"`
		Total         int               `json:"total"`
		FilteredCount int               `json:"filtered_count"`
	} `json:"EPISODES"`
}
