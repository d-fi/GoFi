package types

type LocalesType map[string]struct {
	Name string `json:"name"`
}

type ArtistType struct {
	ART_ID              string      `json:"ART_ID"`
	ROLE_ID             string      `json:"ROLE_ID"`
	ARTISTS_SONGS_ORDER string      `json:"ARTISTS_SONGS_ORDER"`
	ART_NAME            string      `json:"ART_NAME"`
	ARTIST_IS_DUMMY     bool        `json:"ARTIST_IS_DUMMY"`
	ART_PICTURE         string      `json:"ART_PICTURE"`
	RANK                string      `json:"RANK"`
	LOCALES             LocalesType `json:"LOCALES,omitempty"`
	__TYPE__            string      `json:"__TYPE__"`
}

type ArtistInfoTypeMinimal struct {
	ART_ID          string   `json:"ART_ID"`
	ART_NAME        string   `json:"ART_NAME"`
	ART_PICTURE     string   `json:"ART_PICTURE"`
	NB_FAN          int      `json:"NB_FAN"`
	LOCALES         []string `json:"LOCALES"`
	ARTIST_IS_DUMMY bool     `json:"ARTIST_IS_DUMMY"`
	__TYPE__        string   `json:"__TYPE__"`
}

type ArtistInfoType struct {
	ART_ID          string  `json:"ART_ID"`
	ART_NAME        string  `json:"ART_NAME"`
	ARTIST_IS_DUMMY bool    `json:"ARTIST_IS_DUMMY"`
	ART_PICTURE     string  `json:"ART_PICTURE"`
	FACEBOOK        *string `json:"FACEBOOK,omitempty"`
	NB_FAN          int     `json:"NB_FAN"`
	TWITTER         *string `json:"TWITTER,omitempty"`
	__TYPE__        string  `json:"__TYPE__"`
}
