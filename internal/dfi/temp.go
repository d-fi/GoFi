package dfi

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

const defaultDownloadTempMaxAge = time.Hour

func CleanupStaleDownloadTemps(dir string, maxAge time.Duration) (int, error) {
	if dir == "" {
		dir = "."
	}
	if maxAge <= 0 {
		maxAge = defaultDownloadTempMaxAge
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, err
	}

	removed := 0
	cutoff := time.Now().Add(-maxAge)
	for _, entry := range entries {
		if entry.IsDir() || !isDownloadTempName(entry.Name()) {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return removed, err
		}
		if info.ModTime().After(cutoff) {
			continue
		}
		if err := os.Remove(filepath.Join(dir, entry.Name())); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return removed, err
		}
		removed++
	}
	return removed, nil
}

func isDownloadTempName(name string) bool {
	parts := strings.Split(name, "_")
	if len(parts) != 4 || parts[0] != "d-fi" {
		return false
	}
	switch parts[1] {
	case "1", "3", "9":
		return parts[2] != "" && parts[3] != ""
	default:
		return false
	}
}
