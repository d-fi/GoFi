package dfi

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/d-fi/GoFi/metadata"
)

type Config struct {
	Concurrency        int          `json:"concurrency"`
	SaveLayout         SaveLayouts  `json:"saveLayout"`
	Playlist           PlaylistConf `json:"playlist"`
	TrackNumber        bool         `json:"trackNumber"`
	FallbackTrack      bool         `json:"fallbackTrack"`
	FallbackQuality    bool         `json:"fallbackQuality"`
	CoverSize          CoverSizes   `json:"coverSize"`
	Cover              CoverConfig  `json:"cover"`
	Cookies            Cookies      `json:"cookies"`
	path               string
	UserConfigLocation string `json:"-"`
}

type SaveLayouts struct {
	Track    string `json:"track"`
	Album    string `json:"album"`
	Artist   string `json:"artist"`
	Playlist string `json:"playlist"`
}

type PlaylistConf struct {
	ResolveFullPath bool `json:"resolveFullPath"`
}

type CoverSizes struct {
	MP3_128 int `json:"128"`
	MP3_320 int `json:"320"`
	FLAC    int `json:"flac"`
}

type CoverConfig struct {
	Mode string `json:"mode"`
}

type Cookies struct {
	ARL string `json:"arl"`
}

func defaultConfig() Config {
	return Config{
		Concurrency: 4,
		SaveLayout: SaveLayouts{
			Track:    "Music/{ALB_TITLE}/{SNG_TITLE}",
			Album:    "Music/{ALB_TITLE}/{SNG_TITLE}",
			Artist:   "Music/{ALB_TITLE}/{SNG_TITLE}",
			Playlist: "Playlist/{TITLE}/{SNG_TITLE}",
		},
		Playlist: PlaylistConf{
			ResolveFullPath: false,
		},
		TrackNumber:     true,
		FallbackTrack:   true,
		FallbackQuality: true,
		CoverSize: CoverSizes{
			MP3_128: 500,
			MP3_320: 500,
			FLAC:    1000,
		},
		Cover: CoverConfig{
			Mode: string(metadata.CoverModeEmbed),
		},
	}
}

func LoadConfig(path string) Config {
	cfg := defaultConfig()
	cfg.path = path

	data, err := os.ReadFile(path)
	if err != nil {
		return cfg
	}

	var user Config
	if err := json.Unmarshal(data, &user); err != nil {
		fmt.Fprintln(os.Stderr, failure("Unable to parse config: "+path))
		fmt.Fprintln(os.Stderr, note(err.Error()))
		fmt.Fprintln(os.Stderr, warn("Falling back to default config"))
		return cfg
	}

	mergeConfig(&cfg, user)
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err == nil {
		if value, ok := raw["trackNumber"]; ok {
			_ = json.Unmarshal(value, &cfg.TrackNumber)
		}
		if value, ok := raw["fallbackTrack"]; ok {
			_ = json.Unmarshal(value, &cfg.FallbackTrack)
		}
		if value, ok := raw["fallbackQuality"]; ok {
			_ = json.Unmarshal(value, &cfg.FallbackQuality)
		}
		if value, ok := raw["playlist"]; ok {
			var playlistRaw map[string]json.RawMessage
			if err := json.Unmarshal(value, &playlistRaw); err == nil {
				if resolve, ok := playlistRaw["resolveFullPath"]; ok {
					_ = json.Unmarshal(resolve, &cfg.Playlist.ResolveFullPath)
				}
			}
		}
	}
	cfg.UserConfigLocation = path
	return cfg
}

func mergeConfig(cfg *Config, user Config) {
	if user.Concurrency != 0 {
		cfg.Concurrency = user.Concurrency
	}
	if user.SaveLayout.Track != "" {
		cfg.SaveLayout.Track = user.SaveLayout.Track
	}
	if user.SaveLayout.Album != "" {
		cfg.SaveLayout.Album = user.SaveLayout.Album
	}
	if user.SaveLayout.Artist != "" {
		cfg.SaveLayout.Artist = user.SaveLayout.Artist
	}
	if user.SaveLayout.Playlist != "" {
		cfg.SaveLayout.Playlist = user.SaveLayout.Playlist
	}
	cfg.Playlist.ResolveFullPath = user.Playlist.ResolveFullPath
	if user.TrackNumber {
		cfg.TrackNumber = user.TrackNumber
	}
	if user.FallbackTrack {
		cfg.FallbackTrack = user.FallbackTrack
	}
	if user.FallbackQuality {
		cfg.FallbackQuality = user.FallbackQuality
	}
	if user.CoverSize.MP3_128 != 0 {
		cfg.CoverSize.MP3_128 = user.CoverSize.MP3_128
	}
	if user.CoverSize.MP3_320 != 0 {
		cfg.CoverSize.MP3_320 = user.CoverSize.MP3_320
	}
	if user.CoverSize.FLAC != 0 {
		cfg.CoverSize.FLAC = user.CoverSize.FLAC
	}
	if user.Cover.Mode != "" {
		cfg.Cover.Mode = NormalizeCoverMode(user.Cover.Mode)
	}
	if user.Cookies.ARL != "" {
		cfg.Cookies.ARL = user.Cookies.ARL
	}
}

func (cfg *Config) Set(key string, value any) error {
	switch key {
	case "cookies.arl":
		cfg.Cookies.ARL = fmt.Sprintf("%v", value)
	case "concurrency":
		if v, ok := value.(int); ok {
			cfg.Concurrency = v
		}
	case "cover.mode":
		cfg.Cover.Mode = NormalizeCoverMode(fmt.Sprintf("%v", value))
	default:
		return fmt.Errorf("unsupported config key: %s", key)
	}
	return cfg.Save()
}

func NormalizeCoverMode(mode string) string {
	return string(metadata.NormalizeCoverMode(metadata.CoverMode(mode)))
}

func (cfg Config) Save() error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(cfg.path, append(data, '\n'), 0644)
}

func (cfg Config) Layout(linkType string) string {
	switch strings.ToLower(linkType) {
	case "album":
		return cfg.SaveLayout.Album
	case "artist":
		return cfg.SaveLayout.Artist
	case "playlist":
		return cfg.SaveLayout.Playlist
	default:
		return cfg.SaveLayout.Track
	}
}
