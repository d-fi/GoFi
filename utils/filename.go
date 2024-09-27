package utils

import (
	"regexp"
	"runtime"
	"strings"

	"github.com/d-fi/GoFi/logger"
)

// SanitizeFileName replaces characters that are not allowed in file names.
func SanitizeFileName(name string) string {
	logger.Debug("Sanitizing file name: %s", name)

	// Define a regex to match invalid characters
	reg := regexp.MustCompile(`[\/\?<>\\:\*\|"\x00-\x1F]`)

	// Replace invalid characters with underscores
	safeName := reg.ReplaceAllString(name, "_")
	logger.Debug("Replaced invalid characters in file name: %s", safeName)

	// Trim leading and trailing periods and spaces which are also problematic
	safeName = strings.Trim(safeName, ". ")
	logger.Debug("Trimmed leading and trailing periods and spaces: %s", safeName)

	// On Windows, filenames cannot end with a period
	if strings.HasSuffix(safeName, ".") && strings.ToLower(runtime.GOOS) == "windows" {
		safeName = safeName[:len(safeName)-1] + "_"
		logger.Debug("Adjusted filename for Windows: %s", safeName)
	}

	return safeName
}
