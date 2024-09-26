package utils

import (
	"regexp"
	"strings"
)

// SanitizeFileName replaces characters that are not allowed in file names.
func SanitizeFileName(name string) string {
	// Define a regex to match invalid characters
	reg := regexp.MustCompile(`[<>:"/\\|?*\x00-\x1F]`)
	// Replace invalid characters with underscores
	safeName := reg.ReplaceAllString(name, "_")
	// Trim trailing periods and spaces which are also problematic
	safeName = strings.TrimRight(safeName, ". ")
	return safeName
}
