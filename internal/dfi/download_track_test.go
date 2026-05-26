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

func TestWritePlaylistFileRelative(t *testing.T) {
	dir := t.TempDir()
	albumDir := filepath.Join(dir, "album")
	if err := os.MkdirAll(albumDir, 0755); err != nil {
		t.Fatal(err)
	}
	files := []string{
		filepath.Join(albumDir, "02 - B.mp3"),
		filepath.Join(albumDir, "01 - A.mp3"),
	}

	path, err := WritePlaylistFile(map[string]any{"TITLE": "My Playlist"}, files, false)
	if err != nil {
		t.Fatal(err)
	}
	if path != filepath.Join(albumDir, "My Playlist.m3u8") {
		t.Fatalf("playlist path = %q", path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := "#EXTM3U\n01 - A.mp3\n02 - B.mp3"
	if string(data) != want {
		t.Fatalf("playlist content = %q, want %q", data, want)
	}
}
