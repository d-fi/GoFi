package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/d-fi/GoFi/types"
)

func TestParseQuality(t *testing.T) {
	tests := []struct {
		value string
		code  int
		label string
	}{
		{value: "", code: 3, label: "320"},
		{value: "128", code: 1, label: "128"},
		{value: "flac", code: 9, label: "flac"},
	}

	for _, tt := range tests {
		t.Run(tt.label, func(t *testing.T) {
			code, label, err := parseQuality(tt.value)
			if err != nil {
				t.Fatal(err)
			}
			if code != tt.code || label != tt.label {
				t.Fatalf("parseQuality(%q) = %d/%s, want %d/%s", tt.value, code, label, tt.code, tt.label)
			}
		})
	}
}

func TestSelectTracks(t *testing.T) {
	tracks := []types.TrackType{
		{SongType: types.SongType{SNG_ID: "1"}},
		{SongType: types.SongType{SNG_ID: "2"}},
		{SongType: types.SongType{SNG_ID: "3"}},
	}
	selected := selectTracks(tracks, []int{2, 0, 2, 9})
	if len(selected) != 2 {
		t.Fatalf("len(selected) = %d, want 2", len(selected))
	}
	if selected[0].SNG_ID != "3" || selected[1].SNG_ID != "1" {
		t.Fatalf("selected ids = %s/%s, want 3/1", selected[0].SNG_ID, selected[1].SNG_ID)
	}
}

func TestConfigHandlers(t *testing.T) {
	path := filepath.Join(t.TempDir(), "d-fi.config.json")
	server := NewServer(Options{ConfigPath: path})

	req := httptest.NewRequest(http.MethodGet, "/api/config", nil)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/config status = %d", rec.Code)
	}

	var config map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &config); err != nil {
		t.Fatal(err)
	}
	body := config["config"]
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	var cfg map[string]any
	if err := json.Unmarshal(raw, &cfg); err != nil {
		t.Fatal(err)
	}
	cfg["concurrency"] = float64(2)
	bodyBytes, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}

	req = httptest.NewRequest(http.MethodPut, "/api/config", bytes.NewReader(bodyBytes))
	rec = httptest.NewRecorder()
	server.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("PUT /api/config status = %d body=%s", rec.Code, rec.Body.String())
	}
	if got := server.currentConfig().Concurrency; got != 2 {
		t.Fatalf("Concurrency = %d, want 2", got)
	}
}
