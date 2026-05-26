package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"
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
}

func TestClearJobsKeepsActiveJobs(t *testing.T) {
	server := NewServer(Options{ConfigPath: filepath.Join(t.TempDir(), "d-fi.config.json")})
	now := time.Now()
	server.jobs[1] = &downloadJob{ID: 1, Status: "done", CreatedAt: now, UpdatedAt: now}
	server.jobs[2] = &downloadJob{ID: 2, Status: "error", CreatedAt: now, UpdatedAt: now}
	server.jobs[3] = &downloadJob{ID: 3, Status: "queued", CreatedAt: now, UpdatedAt: now}
	server.jobs[4] = &downloadJob{ID: 4, Status: "running", CreatedAt: now, UpdatedAt: now}

	req := httptest.NewRequest(http.MethodDelete, "/api/jobs", nil)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("DELETE /api/jobs status = %d", rec.Code)
	}
	if server.jobs[1] != nil || server.jobs[2] != nil {
		t.Fatal("inactive jobs were not cleared")
	}
	if server.jobs[3] == nil || server.jobs[4] == nil {
		t.Fatal("active jobs should be kept")
	}
}
