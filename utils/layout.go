package utils

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/d-fi/GoFi/logger"
)

// SaveLayoutProps holds the parameters required for the SaveLayout function.
type SaveLayoutProps struct {
	Track                map[string]interface{}
	Album                map[string]interface{}
	Path                 string
	MinimumIntegerDigits int
	TrackNumber          bool
}

// SaveLayout formats the file path using placeholders and metadata from the track and album.
func SaveLayout(props SaveLayoutProps) string {
	logger.Debug("Starting save layout formatting for path: %s", props.Path)

	// Clone album info to avoid modifying the original map
	albumInfo := make(map[string]interface{})
	for k, v := range props.Album {
		albumInfo[k] = v
	}

	// Use relative path if it starts with '{'
	if strings.HasPrefix(props.Path, "{") {
		props.Path = "./" + props.Path
		logger.Debug("Updated path to be relative: %s", props.Path)
	}

	// Find keys inside {}
	re := regexp.MustCompile(`\{([^}]*)\}`)
	matches := re.FindAllStringSubmatch(props.Path, -1)

	for _, match := range matches {
		key := match[1]
		logger.Debug("Processing key: %s", key)

		valueAlbum, okAlbum := albumInfo[key]
		valueTrack, okTrack := props.Track[key]

		var value string
		if okAlbum {
			value = fmt.Sprintf("%v", valueAlbum)
			logger.Debug("Found value from album: %s = %s", key, value)
		} else if okTrack {
			value = fmt.Sprintf("%v", valueTrack)
			logger.Debug("Found value from track: %s = %s", key, value)
		}

		if key == "TRACK_NUMBER" || key == "TRACK_POSITION" || key == "NO_TRACK_NUMBER" {
			if value != "" {
				num, _ := strconv.Atoi(value)
				props.Path = strings.Replace(props.Path, "{"+key+"}", fmt.Sprintf("%0*d", props.MinimumIntegerDigits, num), -1)
				logger.Debug("Formatted track number for key %s: %s", key, props.Path)
			} else {
				props.Path = strings.Replace(props.Path, "{"+key+"}", "", -1)
				logger.Debug("Key %s had no value; replaced with empty string.", key)
			}
			props.TrackNumber = false
		} else {
			props.Path = strings.Replace(props.Path, "{"+key+"}", SanitizeFileName(value), -1)
			logger.Debug("Replaced key %s with sanitized value: %s", key, props.Path)
		}
	}

	if props.TrackNumber {
		if trackNum, exists := props.Track["TRACK_NUMBER"]; exists {
			trackNumber := fmt.Sprintf("%0*d", props.MinimumIntegerDigits, trackNum)
			props.Path = filepath.Join(filepath.Dir(props.Path), trackNumber+" - "+filepath.Base(props.Path))
			logger.Debug("Appended track number to path: %s", props.Path)
		}
	}

	// Remove any remaining problematic characters
	finalPath := strings.Trim(regexp.MustCompile(`[?%*|"<>]`).ReplaceAllString(props.Path, ""), " ")
	logger.Debug("Final sanitized path: %s", finalPath)
	return finalPath
}
