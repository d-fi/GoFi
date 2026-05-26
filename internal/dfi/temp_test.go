package dfi

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCleanupStaleDownloadTemps(t *testing.T) {
	dir := t.TempDir()
	stale := filepath.Join(dir, "d-fi_3_2202736507_99573fea3e0593ead564fff9eab9edcf")
	fresh := filepath.Join(dir, "d-fi_9_2202736508_99573fea3e0593ead564fff9eab9edcf")
	other := filepath.Join(dir, "d-fi.config.json")

	for _, path := range []string{stale, fresh, other} {
		if err := os.WriteFile(path, []byte("temp"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	oldTime := time.Now().Add(-2 * time.Hour)
	if err := os.Chtimes(stale, oldTime, oldTime); err != nil {
		t.Fatal(err)
	}

	removed, err := CleanupStaleDownloadTemps(dir, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if removed != 1 {
		t.Fatalf("removed = %d, want 1", removed)
	}
	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Fatalf("stale temp still exists or stat failed with %v", err)
	}
	for _, path := range []string{fresh, other} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("%s should remain: %v", filepath.Base(path), err)
		}
	}
}

func TestIsDownloadTempName(t *testing.T) {
	tests := map[string]bool{
		"d-fi_1_123_md5":       true,
		"d-fi_3_123_md5":       true,
		"d-fi_9_123_simulate":  true,
		"d-fi_2_123_md5":       false,
		"d-fi_3_123":           false,
		"d-fi.config.json":     false,
		"d-fi_3_123_md5_extra": false,
		"d-fi_3__md5":          false,
		"d-fi_3_123_":          false,
		"other-d-fi_3_123_md5": false,
	}
	for name, want := range tests {
		if got := isDownloadTempName(name); got != want {
			t.Fatalf("isDownloadTempName(%q) = %v, want %v", name, got, want)
		}
	}
}
