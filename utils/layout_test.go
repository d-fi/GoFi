package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSaveLayout(t *testing.T) {
	tests := []struct {
		name     string
		props    SaveLayoutProps
		expected string
	}{
		{
			name: "Basic_placeholders",
			props: SaveLayoutProps{
				Track: map[string]interface{}{
					"TRACK_NUMBER": 1,
					"TITLE":        "Song Title",
				},
				Album: map[string]interface{}{
					"ARTIST":    "Artist Name",
					"ALB_TITLE": "Album Name",
				},
				Path:                 "{ARTIST}/{ALB_TITLE}/{TRACK_NUMBER} - {TITLE}.mp3",
				MinimumIntegerDigits: 2,
				TrackNumber:          false,
			},
			expected: "Artist Name/Album Name/01 - Song Title.mp3",
		},
		{
			name: "Missing_keys",
			props: SaveLayoutProps{
				Track: map[string]interface{}{
					"TITLE": "Song Title",
				},
				Album: map[string]interface{}{
					"ARTIST": "Artist Name",
				},
				Path:                 "{ARTIST}/{ALB_TITLE}/{TRACK_NUMBER} - {TITLE}.mp3",
				MinimumIntegerDigits: 2,
				TrackNumber:          false,
			},
			expected: "Artist Name/ - Song Title.mp3",
		},
		{
			name: "Special_characters",
			props: SaveLayoutProps{
				Track: map[string]interface{}{
					"TRACK_NUMBER": 1,
					"TITLE":        "Song:Title*With|Special<>Chars?",
				},
				Album: map[string]interface{}{
					"ARTIST":    "Artist/Name\\With\"Special:Chars",
					"ALB_TITLE": "Album|Name*With?Special<>Chars",
				},
				Path:                 "{ARTIST}/{ALB_TITLE}/{TRACK_NUMBER} - {TITLE}.mp3",
				MinimumIntegerDigits: 2,
				TrackNumber:          false,
			},
			expected: "Artist_Name_With_Special_Chars/Album_Name_With_Special__Chars/01 - Song_Title_With_Special__Chars_.mp3",
		},
		{
			name: "TrackNumber_flag_true",
			props: SaveLayoutProps{
				Track: map[string]interface{}{
					"TRACK_NUMBER": 5,
					"TITLE":        "Song Title",
				},
				Album: map[string]interface{}{
					"ARTIST":    "Artist Name",
					"ALB_TITLE": "Album Name",
				},
				Path:                 "{ARTIST}/{ALB_TITLE}/{TITLE}.mp3",
				MinimumIntegerDigits: 2,
				TrackNumber:          true,
			},
			expected: "Artist Name/Album Name/05 - Song Title.mp3",
		},
		{
			name: "Relative_path",
			props: SaveLayoutProps{
				Track: map[string]interface{}{
					"TRACK_NUMBER": 3,
					"TITLE":        "Song Title",
				},
				Album: map[string]interface{}{
					"ARTIST":    "Artist Name",
					"ALB_TITLE": "Album Name",
				},
				Path:                 "{ARTIST}/{TITLE}.mp3",
				MinimumIntegerDigits: 2,
				TrackNumber:          true,
			},
			expected: "Artist Name/03 - Song Title.mp3",
		},
		{
			name: "No_placeholders",
			props: SaveLayoutProps{
				Track:                nil,
				Album:                nil,
				Path:                 "static/path/filename.mp3",
				MinimumIntegerDigits: 2,
				TrackNumber:          false,
			},
			expected: "static/path/filename.mp3",
		},
		{
			name: "Invalid_track_number",
			props: SaveLayoutProps{
				Track: map[string]interface{}{
					"TRACK_NUMBER": "invalid",
					"TITLE":        "Song Title",
				},
				Album: map[string]interface{}{
					"ARTIST":    "Artist Name",
					"ALB_TITLE": "Album Name",
				},
				Path:                 "{ARTIST}/{ALB_TITLE}/{TRACK_NUMBER} - {TITLE}.mp3",
				MinimumIntegerDigits: 2,
				TrackNumber:          false,
			},
			expected: "Artist Name/Album Name/00 - Song Title.mp3",
		},
		{
			name: "Disc_number_adjustment",
			props: SaveLayoutProps{
				Track: map[string]interface{}{
					"TRACK_NUMBER": 1,
					"TITLE":        "Song Title",
					"DISK_NUMBER":  2,
				},
				Album: map[string]interface{}{
					"ARTIST":      "Artist Name",
					"ALB_TITLE":   "Album Name",
					"NUMBER_DISK": 3,
				},
				Path:                 "{ARTIST}/{ALB_TITLE}/{TRACK_NUMBER} - {TITLE}.mp3",
				MinimumIntegerDigits: 2,
				TrackNumber:          false,
			},
			expected: "Artist Name/Album Name (Disc 02)/01 - Song Title.mp3",
		},
		{
			name: "Nested_key_access",
			props: SaveLayoutProps{
				Track: map[string]interface{}{
					"INFO": map[string]interface{}{
						"TRACK_NUMBER": 4,
					},
					"TITLE": "Song Title",
				},
				Album: map[string]interface{}{
					"ARTIST": "Artist Name",
				},
				Path:                 "{ARTIST}/{INFO.TRACK_NUMBER} - {TITLE}.mp3",
				MinimumIntegerDigits: 2,
				TrackNumber:          false,
			},
			expected: "Artist Name/04 - Song Title.mp3",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := SaveLayout(test.props)
			assert.Equal(t, test.expected, result, "Test '%s' failed", test.name)
		})
	}
}
