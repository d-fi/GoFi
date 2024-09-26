package types

import (
	"encoding/json"
	"fmt"
)

// LocalesType represents locales with their names, where the key is the locale code.
type LocalesType map[string]struct {
	Name string `json:"name"` // The localized name, e.g., 'Daft Punk', 'ダフトパンク' in different languages
}

// UnmarshalJSON for LocalesType allows dynamic handling of multiple structures: objects, empty arrays, etc.
func (lt *LocalesType) UnmarshalJSON(data []byte) error {
	// Attempt to unmarshal directly into the expected map structure
	var mapData map[string]struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(data, &mapData); err == nil {
		*lt = LocalesType(mapData)
		return nil
	}

	// Handle empty array or empty object
	if string(data) == "[]" || string(data) == "{}" {
		*lt = LocalesType{}
		return nil
	}

	// If parsing fails, return an error
	return fmt.Errorf("failed to unmarshal LocalesType: %s", string(data))
}

// ArtistType represents detailed information about an artist, including their role and ranking.
type ArtistType struct {
	ART_ID              string      `json:"ART_ID"`              // '27'
	ROLE_ID             string      `json:"ROLE_ID"`             // '0'
	ARTISTS_SONGS_ORDER string      `json:"ARTISTS_SONGS_ORDER"` // '0'
	ART_NAME            string      `json:"ART_NAME"`            // 'Daft Punk'
	ARTIST_IS_DUMMY     bool        `json:"ARTIST_IS_DUMMY"`     // false
	ART_PICTURE         string      `json:"ART_PICTURE"`         // 'f2bc007e9133c946ac3c3907ddc5d2ea'
	RANK                string      `json:"RANK"`                // '836071'
	LOCALES             LocalesType `json:"LOCALES,omitempty"`   // Optional locales with translations
	TYPE_INTERNAL       string      `json:"__TYPE__"`            // 'artist'
}

// ArtistInfoTypeMinimal represents minimal information about an artist, including their name and fan count.
type ArtistInfoTypeMinimal struct {
	ART_ID          string   `json:"ART_ID"`          // Artist ID, e.g., '27'
	ART_NAME        string   `json:"ART_NAME"`        // Artist name, e.g., 'Daft Punk'
	ART_PICTURE     string   `json:"ART_PICTURE"`     // Artist picture URL, e.g., 'f2bc007e9133c946ac3c3907ddc5d2ea'
	NB_FAN          int      `json:"NB_FAN"`          // Number of fans, e.g., 7140516
	LOCALES         []string `json:"LOCALES"`         // Locales, can be an empty array or specific locale strings
	ARTIST_IS_DUMMY bool     `json:"ARTIST_IS_DUMMY"` // Indicates if the artist is a dummy, e.g., false
	TYPE_INTERNAL   string   `json:"__TYPE__"`        // Internal type, e.g., 'artist'
}

// ArtistInfoType represents detailed information about an artist, including social media links.
type ArtistInfoType struct {
	ART_ID          string  `json:"ART_ID"`             // '293585'
	ART_NAME        string  `json:"ART_NAME"`           // 'Avicii'
	ARTIST_IS_DUMMY bool    `json:"ARTIST_IS_DUMMY"`    // Indicates if the artist is a dummy, e.g., false
	ART_PICTURE     string  `json:"ART_PICTURE"`        // '82e214b0cb39316f4a12a082fded54f6'
	FACEBOOK        *string `json:"FACEBOOK,omitempty"` // Optional Facebook URL, e.g., 'https://www.facebook.com/avicii?fref=ts'
	NB_FAN          int     `json:"NB_FAN"`             // Number of fans, e.g., 7140516
	TWITTER         *string `json:"TWITTER,omitempty"`  // Optional Twitter URL, e.g., 'https://twitter.com/Avicii'
	TYPE_INTERNAL   string  `json:"__TYPE__"`           // 'artist'
}
