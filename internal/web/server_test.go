package web

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/d-fi/GoFi/types"
)

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
	if got := server.currentConfig().Cover.Mode; got != "embed" {
		t.Fatalf("Cover.Mode = %q, want embed", got)
	}
	if got := server.currentConfig().Cover.FileName; got != "cover.jpg" {
		t.Fatalf("Cover.FileName = %q, want cover.jpg", got)
	}
}

func TestStaticAssetHandlers(t *testing.T) {
	server := NewServer(Options{ConfigPath: filepath.Join(t.TempDir(), "d-fi.config.json")})

	tests := map[string]string{
		"/":          "text/html; charset=utf-8",
		"/style.css": "text/css; charset=utf-8",
		"/script.js": "application/javascript; charset=utf-8",
	}
	for path, contentType := range tests {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("GET %s status = %d", path, rec.Code)
		}
		if got := rec.Header().Get("Content-Type"); got != contentType {
			t.Fatalf("GET %s Content-Type = %q, want %q", path, got, contentType)
		}
		if rec.Body.Len() == 0 {
			t.Fatalf("GET %s returned an empty body", path)
		}
	}
}

func TestConfigUpdatePreservesCoverSettingsWhenMissing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "d-fi.config.json")
	server := NewServer(Options{ConfigPath: path})
	server.cfg.Cover.Mode = "file"
	server.cfg.Cover.FileName = "folder.jpg"

	body := []byte(`{
		"concurrency": 3,
		"trackNumber": true,
		"fallbackTrack": true,
		"fallbackQuality": true,
		"saveLayout": {
			"track": "Music/{ALB_TITLE}/{SNG_TITLE}",
			"album": "Music/{ALB_TITLE}/{SNG_TITLE}",
			"artist": "Music/{ALB_TITLE}/{SNG_TITLE}",
			"playlist": "Playlist/{TITLE}/{SNG_TITLE}"
		},
		"playlist": {"resolveFullPath": false},
		"coverSize": {"128": 500, "320": 500, "flac": 1000},
		"cookies": {"arl": ""}
	}`)
	req := httptest.NewRequest(http.MethodPut, "/api/config", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("PUT /api/config status = %d body=%s", rec.Code, rec.Body.String())
	}
	if got := server.currentConfig().Cover.Mode; got != "file" {
		t.Fatalf("Cover.Mode = %q, want file", got)
	}
	if got := server.currentConfig().Cover.FileName; got != "folder.jpg" {
		t.Fatalf("Cover.FileName = %q, want folder.jpg", got)
	}
}

func TestClearJobsKeepsActiveJobs(t *testing.T) {
	server := NewServer(Options{ConfigPath: filepath.Join(t.TempDir(), "d-fi.config.json")})
	now := time.Now()
	server.jobs[1] = &downloadJob{ID: 1, Status: "done", CreatedAt: now, UpdatedAt: now}
	server.jobs[2] = &downloadJob{ID: 2, Status: "error", CreatedAt: now, UpdatedAt: now}
	server.jobs[3] = &downloadJob{ID: 3, Status: "queued", CreatedAt: now, UpdatedAt: now}
	server.jobs[4] = &downloadJob{ID: 4, Status: "running", CreatedAt: now, UpdatedAt: now}
	server.jobs[5] = &downloadJob{ID: 5, Status: "canceling", CreatedAt: now, UpdatedAt: now}

	req := httptest.NewRequest(http.MethodDelete, "/api/jobs", nil)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("DELETE /api/jobs status = %d", rec.Code)
	}
	if server.jobs[1] != nil || server.jobs[2] != nil {
		t.Fatal("inactive jobs were not cleared")
	}
	if server.jobs[3] == nil || server.jobs[4] == nil || server.jobs[5] == nil {
		t.Fatal("active jobs should be kept")
	}
}

func TestCancelJobMarksJobCanceling(t *testing.T) {
	server := NewServer(Options{ConfigPath: filepath.Join(t.TempDir(), "d-fi.config.json")})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	now := time.Now()
	server.jobs[1] = &downloadJob{
		ID:        1,
		Status:    "running",
		CreatedAt: now,
		UpdatedAt: now,
		cancel:    cancel,
	}

	req := httptest.NewRequest(http.MethodPost, "/api/jobs/1/cancel", nil)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("POST /api/jobs/1/cancel status = %d body=%s", rec.Code, rec.Body.String())
	}
	if got := server.jobs[1].Status; got != "canceling" {
		t.Fatalf("Status = %q, want canceling", got)
	}
	if err := ctx.Err(); err == nil {
		t.Fatal("cancel function was not called")
	}
}

func TestLayoutFieldsIncludesAlwaysAndCurrentResponseFields(t *testing.T) {
	fields := layoutFields("playlist", map[string]any{
		"TITLE":      "My Playlist",
		"DATE_ADD":   "2026-05-29",
		"nested":     map[string]any{"value": "x"},
		"empty":      "",
		"empty_list": []any{},
	}, []types.TrackType{
		{
			SongType: types.SongType{
				ALB_TITLE:    "Discovery",
				ART_NAME:     "Daft Punk",
				SNG_TITLE:    "One More Time",
				DATE_START:   "2001-03-07",
				TRACK_NUMBER: 1,
				STATUS:       0,
			},
		},
	})

	if hasLayoutField(fields.Always, "RELEASE_YEAR") {
		t.Fatal("always fields should not include response-dependent RELEASE_YEAR")
	}
	if !hasLayoutField(fields.Always, "TITLE") {
		t.Fatal("playlist always fields should include TITLE")
	}
	if !hasLayoutField(fields.Current, "DATE_ADD") {
		t.Fatal("current fields should include info response fields")
	}
	if !hasLayoutField(fields.Current, "nested.value") {
		t.Fatal("current fields should include nested response fields")
	}
	if !hasLayoutField(fields.Current, "SNG_TITLE") {
		t.Fatal("current fields should include track response fields")
	}
	if !hasLayoutField(fields.Current, "RELEASE_YEAR") {
		t.Fatal("current fields should include derived release fields when a release date exists")
	}
	if hasLayoutField(fields.Current, "empty") {
		t.Fatal("current fields should skip empty response fields")
	}
	if hasLayoutField(fields.Current, "STATUS") {
		t.Fatal("current fields should skip zero response fields")
	}
}

func hasLayoutField(fields []layoutField, key string) bool {
	for _, field := range fields {
		if field.Key == key {
			return true
		}
	}
	return false
}
