package dfi

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/d-fi/GoFi/download"
)

func TestDownloadToTempRestartsWhenRangeIgnored(t *testing.T) {
	dir := t.TempDir()
	tmpFile := filepath.Join(dir, "d-fi_partial")
	if err := os.WriteFile(tmpFile, []byte("partial-"), 0644); err != nil {
		t.Fatal(err)
	}

	var sawRange bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Range") != "" {
			sawRange = true
		}
		_, _ = w.Write([]byte("complete"))
	}))
	defer server.Close()

	err := downloadToTemp(context.Background(), &download.TrackDownloadUrl{
		TrackUrl:    server.URL,
		IsEncrypted: false,
		FileSize:    int64(len("complete")),
	}, tmpFile, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !sawRange {
		t.Fatal("expected first request to include Range header")
	}

	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "complete" {
		t.Fatalf("temp file = %q, want complete", data)
	}
}

func TestDownloadToTempAppendsPartialContent(t *testing.T) {
	dir := t.TempDir()
	tmpFile := filepath.Join(dir, "d-fi_partial")
	if err := os.WriteFile(tmpFile, []byte("partial-"), 0644); err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Range"); got != "bytes=8-" {
			http.Error(w, fmt.Sprintf("unexpected range %q", got), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusPartialContent)
		_, _ = w.Write([]byte("complete"))
	}))
	defer server.Close()

	err := downloadToTemp(context.Background(), &download.TrackDownloadUrl{
		TrackUrl:    server.URL,
		IsEncrypted: false,
		FileSize:    int64(len("partial-complete")),
	}, tmpFile, nil)
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "partial-complete" {
		t.Fatalf("temp file = %q, want partial-complete", data)
	}
}
