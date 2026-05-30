package dfi

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigDefaults(t *testing.T) {
	cfg := LoadConfig(filepath.Join(t.TempDir(), "missing.json"))

	if cfg.Concurrency != 4 {
		t.Fatalf("Concurrency = %d, want 4", cfg.Concurrency)
	}
	if cfg.SaveLayout.Track != "Music/{ALB_TITLE}/{SNG_TITLE}" {
		t.Fatalf("unexpected track layout: %s", cfg.SaveLayout.Track)
	}
	if cfg.SaveLayout.Album != "Music/{ALB_TITLE}/{SNG_TITLE}" {
		t.Fatalf("unexpected album layout: %s", cfg.SaveLayout.Album)
	}
	if !cfg.TrackNumber {
		t.Fatal("TrackNumber should default true")
	}
	if !cfg.FallbackTrack {
		t.Fatal("FallbackTrack should default true")
	}
	if cfg.Cookies.ARL != "" {
		t.Fatal("ARL should not be hardcoded in defaults")
	}
	if cfg.Cover.Mode != "embed" {
		t.Fatalf("Cover.Mode = %q, want embed", cfg.Cover.Mode)
	}
	if cfg.Cover.FileName != "cover.jpg" {
		t.Fatalf("Cover.FileName = %q, want cover.jpg", cfg.Cover.FileName)
	}
}

func TestLoadConfigMergesFalseValues(t *testing.T) {
	path := filepath.Join(t.TempDir(), "d-fi.config.json")
	if err := os.WriteFile(path, []byte(`{
		"concurrency": 2,
		"trackNumber": false,
			"fallbackTrack": false,
			"fallbackQuality": false,
			"cover": {"mode": "file", "fileName": "folder.jpg"},
			"playlist": {"resolveFullPath": true},
			"saveLayout": {"track": "{ART_NAME}/{SNG_TITLE}"}
		}`), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := LoadConfig(path)
	if cfg.Concurrency != 2 {
		t.Fatalf("Concurrency = %d, want 2", cfg.Concurrency)
	}
	if cfg.TrackNumber {
		t.Fatal("TrackNumber should be false from config")
	}
	if cfg.FallbackTrack {
		t.Fatal("FallbackTrack should be false from config")
	}
	if cfg.FallbackQuality {
		t.Fatal("FallbackQuality should be false from config")
	}
	if !cfg.Playlist.ResolveFullPath {
		t.Fatal("ResolveFullPath should be true from config")
	}
	if cfg.SaveLayout.Track != "{ART_NAME}/{SNG_TITLE}" {
		t.Fatalf("unexpected track layout: %s", cfg.SaveLayout.Track)
	}
	if cfg.SaveLayout.Album != "Music/{ALB_TITLE}/{SNG_TITLE}" {
		t.Fatalf("album layout default was not preserved: %s", cfg.SaveLayout.Album)
	}
	if cfg.Cover.Mode != "file" {
		t.Fatalf("Cover.Mode = %q, want file", cfg.Cover.Mode)
	}
	if cfg.Cover.FileName != "folder.jpg" {
		t.Fatalf("Cover.FileName = %q, want folder.jpg", cfg.Cover.FileName)
	}
}

func TestConfigSetARL(t *testing.T) {
	path := filepath.Join(t.TempDir(), "d-fi.config.json")
	cfg := LoadConfig(path)

	if err := cfg.Set("cookies.arl", "abc"); err != nil {
		t.Fatal(err)
	}
	loaded := LoadConfig(path)
	if loaded.Cookies.ARL != "abc" {
		t.Fatalf("ARL = %q, want abc", loaded.Cookies.ARL)
	}
}

func TestConfigSetConcurrency(t *testing.T) {
	path := filepath.Join(t.TempDir(), "d-fi.config.json")
	cfg := LoadConfig(path)

	if err := cfg.Set("concurrency", 9); err != nil {
		t.Fatal(err)
	}
	loaded := LoadConfig(path)
	if loaded.Concurrency != 9 {
		t.Fatalf("Concurrency = %d, want 9", loaded.Concurrency)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	if raw["concurrency"] != float64(9) {
		t.Fatalf("persisted concurrency = %#v, want 9", raw["concurrency"])
	}
}

func TestConfigSetCoverMode(t *testing.T) {
	path := filepath.Join(t.TempDir(), "d-fi.config.json")
	cfg := LoadConfig(path)

	if err := cfg.Set("cover.mode", "both"); err != nil {
		t.Fatal(err)
	}
	loaded := LoadConfig(path)
	if loaded.Cover.Mode != "both" {
		t.Fatalf("Cover.Mode = %q, want both", loaded.Cover.Mode)
	}
}

func TestConfigSetCoverFileName(t *testing.T) {
	path := filepath.Join(t.TempDir(), "d-fi.config.json")
	cfg := LoadConfig(path)

	if err := cfg.Set("cover.fileName", "../Folder.png"); err != nil {
		t.Fatal(err)
	}
	loaded := LoadConfig(path)
	if loaded.Cover.FileName != "Folder.jpg" {
		t.Fatalf("Cover.FileName = %q, want Folder.jpg", loaded.Cover.FileName)
	}
}

func TestResolveARLPrefersEnv(t *testing.T) {
	t.Setenv("DEEZER_ARL", "env-arl")

	got := resolveARL(Config{Cookies: Cookies{ARL: "config-arl"}})
	if got != "env-arl" {
		t.Fatalf("resolveARL = %q, want env-arl", got)
	}
}

func TestResolveARLFallsBackToConfig(t *testing.T) {
	t.Setenv("DEEZER_ARL", "")

	got := resolveARL(Config{Cookies: Cookies{ARL: " config-arl "}})
	if got != "config-arl" {
		t.Fatalf("resolveARL = %q, want config-arl", got)
	}
}
