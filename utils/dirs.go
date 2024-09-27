package utils

import (
	"log"
	"os"
	"path/filepath"
)

func GetSystemCacheDir(appName string) string {
	var cachePath string

	if xdgCacheHome := os.Getenv("XDG_CACHE_HOME"); xdgCacheHome != "" {
		cachePath = filepath.Join(xdgCacheHome, appName)
	} else if homeDir, err := os.UserHomeDir(); err == nil {
		cachePath = filepath.Join(homeDir, ".cache", appName)
	} else {
		cachePath = filepath.Join(os.TempDir(), appName)
	}

	if err := os.MkdirAll(cachePath, 0755); err != nil {
		log.Fatalf("Failed to create cache directory: %v", err)
	}

	return cachePath
}
