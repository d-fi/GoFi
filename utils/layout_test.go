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
				Track: map[string]any{
					"TRACK_NUMBER": 1,
					"TITLE":        "Song Title",
				},
				Album: map[string]any{
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
				Track: map[string]any{
					"TITLE": "Song Title",
				},
				Album: map[string]any{
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
				Track: map[string]any{
					"TRACK_NUMBER": 1,
					"TITLE":        "Song:Title*With|Special<>Chars?",
				},
				Album: map[string]any{
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
				Track: map[string]any{
					"TRACK_NUMBER": 5,
					"TITLE":        "Song Title",
				},
				Album: map[string]any{
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
				Track: map[string]any{
					"TRACK_NUMBER": 3,
					"TITLE":        "Song Title",
				},
				Album: map[string]any{
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
				Track: map[string]any{
					"TRACK_NUMBER": "invalid",
					"TITLE":        "Song Title",
				},
				Album: map[string]any{
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
				Track: map[string]any{
					"TRACK_NUMBER": 1,
					"TITLE":        "Song Title",
					"DISK_NUMBER":  2,
				},
				Album: map[string]any{
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
			name: "Disk_folder_placeholder_keeps_album_title_unchanged",
			props: SaveLayoutProps{
				Track: map[string]any{
					"TRACK_NUMBER": 1,
					"TITLE":        "Song Title",
					"DISK_NUMBER":  2,
				},
				Album: map[string]any{
					"ARTIST":      "Artist Name",
					"ALB_TITLE":   "Album Name",
					"NUMBER_DISK": 3,
				},
				Path:                 "{ARTIST}/{ALB_TITLE}/{DISK_FOLDER}/{TRACK_NUMBER} - {TITLE}.mp3",
				MinimumIntegerDigits: 2,
				TrackNumber:          false,
			},
			expected: "Artist Name/Album Name/CD2/01 - Song Title.mp3",
		},
		{
			name: "Disk_folder_fallback_placeholder_keeps_album_title_unchanged",
			props: SaveLayoutProps{
				Track: map[string]any{
					"TRACK_NUMBER": 1,
					"TITLE":        "Song Title",
					"DISK_NUMBER":  2,
				},
				Album: map[string]any{
					"ARTIST":      "Artist Name",
					"ALB_TITLE":   "Album Name",
					"NUMBER_DISK": 3,
				},
				Path:                 "{ARTIST}/{ALB_TITLE}/{DISK_FOLDER|DISK_NUMBER}/{TRACK_NUMBER} - {TITLE}.mp3",
				MinimumIntegerDigits: 2,
				TrackNumber:          false,
			},
			expected: "Artist Name/Album Name/CD2/01 - Song Title.mp3",
		},
		{
			name: "Single_disc_album_omits_disk_folder",
			props: SaveLayoutProps{
				Track: map[string]any{
					"TRACK_NUMBER": 1,
					"TITLE":        "Song Title",
					"DISK_NUMBER":  1,
				},
				Album: map[string]any{
					"ARTIST":      "Artist Name",
					"ALB_TITLE":   "Album Name",
					"NUMBER_DISK": 1,
				},
				Path:                 "{ARTIST}/{ALB_TITLE}/{DISK_FOLDER}/{TRACK_NUMBER} - {TITLE}.mp3",
				MinimumIntegerDigits: 2,
				TrackNumber:          false,
			},
			expected: "Artist Name/Album Name/01 - Song Title.mp3",
		},
		{
			name: "Nested_key_access",
			props: SaveLayoutProps{
				Track: map[string]any{
					"INFO": map[string]any{
						"TRACK_NUMBER": 4,
					},
					"TITLE": "Song Title",
				},
				Album: map[string]any{
					"ARTIST": "Artist Name",
				},
				Path:                 "{ARTIST}/{INFO.TRACK_NUMBER} - {TITLE}.mp3",
				MinimumIntegerDigits: 2,
				TrackNumber:          false,
			},
			expected: "Artist Name/04 - Song Title.mp3",
		},
		{
			name: "Nested_array_key_access",
			props: SaveLayoutProps{
				Track: map[string]any{
					"ARTISTS": []any{
						map[string]any{"ART_NAME": "Daft Punk"},
					},
					"SNG_CONTRIBUTORS": map[string]any{
						"main_artist": []string{"Daft Punk"},
					},
					"TITLE": "Song Title",
				},
				Album: map[string]any{
					"ALB_TITLE": "Album Name",
				},
				Path:                 "{ARTISTS.0.ART_NAME}/{SNG_CONTRIBUTORS.main_artist.0}/{TITLE}.mp3",
				MinimumIntegerDigits: 2,
				TrackNumber:          false,
			},
			expected: "Daft Punk/Daft Punk/Song Title.mp3",
		},
		{
			name: "Release_date_prefers_original_date_from_album",
			props: SaveLayoutProps{
				Track: map[string]any{
					"TITLE": "Song Title",
				},
				Album: map[string]any{
					"ALB_TITLE":             "Album Name",
					"ORIGINAL_RELEASE_DATE": "1980-01-01",
					"PHYSICAL_RELEASE_DATE": "1990-10-29",
					"DIGITAL_RELEASE_DATE":  "2011-06-10",
				},
				Path:                 "{RELEASE_YEAR}/{RELEASE_DATE}/{ALB_TITLE}/{TITLE}.mp3",
				MinimumIntegerDigits: 2,
				TrackNumber:          false,
			},
			expected: "1980/1980-01-01/Album Name/Song Title.mp3",
		},
		{
			name: "Release_date_aliases_from_public_track_album",
			props: SaveLayoutProps{
				Track: map[string]any{
					"TITLE": "Song Title",
					"album": map[string]any{
						"release_date": "1998-01-20",
					},
				},
				Album:                nil,
				Path:                 "{RELEASE_YEAR}/{TITLE}.mp3",
				MinimumIntegerDigits: 2,
				TrackNumber:          false,
			},
			expected: "1998/Song Title.mp3",
		},
		{
			name: "Fallback_placeholder_uses_first_non_empty_value",
			props: SaveLayoutProps{
				Track: map[string]any{
					"SNG_TITLE": "Song Title",
				},
				Album: map[string]any{
					"ALB_TITLE": "",
					"TITLE":     "Playlist Name",
				},
				Path:                 "{ALB_TITLE|TITLE}/{SNG_TITLE}.mp3",
				MinimumIntegerDigits: 2,
				TrackNumber:          false,
			},
			expected: "Playlist Name/Song Title.mp3",
		},
		{
			name: "Fallback_placeholder_supports_nested_values",
			props: SaveLayoutProps{
				Track: map[string]any{
					"ARTISTS": []any{
						map[string]any{"ART_NAME": "Daft Punk"},
					},
					"SNG_TITLE": "Song Title",
				},
				Album:                nil,
				Path:                 "{ART_NAME|ARTISTS.0.ART_NAME}/{SNG_TITLE}.mp3",
				MinimumIntegerDigits: 2,
				TrackNumber:          false,
			},
			expected: "Daft Punk/Song Title.mp3",
		},
		{
			name: "Fallback_placeholder_formats_track_number",
			props: SaveLayoutProps{
				Track: map[string]any{
					"TRACK_NUMBER": 7,
					"SNG_TITLE":    "Song Title",
				},
				Album:                nil,
				Path:                 "{TRACK_POSITION|TRACK_NUMBER} - {SNG_TITLE}.mp3",
				MinimumIntegerDigits: 2,
				TrackNumber:          false,
			},
			expected: "07 - Song Title.mp3",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := SaveLayout(test.props)
			assert.Equal(t, test.expected, result, "Test '%s' failed", test.name)
		})
	}
}

func TestGetNestedValue(t *testing.T) {
	data := map[string]any{
		"ARTISTS": []any{
			map[string]any{"ART_NAME": "Daft Punk"},
		},
		"SNG_CONTRIBUTORS": map[string]any{
			"main_artist": []string{"Daft Punk"},
		},
	}

	got, ok := GetNestedValue(data, "ARTISTS.0.ART_NAME")
	assert.True(t, ok)
	assert.Equal(t, "Daft Punk", got)

	got, ok = GetNestedValue(data, "SNG_CONTRIBUTORS.main_artist.0")
	assert.True(t, ok)
	assert.Equal(t, "Daft Punk", got)

	_, ok = GetNestedValue(data, "ARTISTS.1.ART_NAME")
	assert.False(t, ok)
}

func TestBestReleaseDate(t *testing.T) {
	got := BestReleaseDate(map[string]any{
		"PHYSICAL_RELEASE_DATE": "1990-10-29",
		"release_date":          "2011-06-10",
	}, map[string]any{
		"DATE_START": "2001-03-07",
	})
	if got != "1990-10-29" {
		t.Fatalf("BestReleaseDate = %q, want 1990-10-29", got)
	}
}
