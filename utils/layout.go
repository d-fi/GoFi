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

// atoiOrZero converts a string to an int, returning 0 if conversion fails.
func atoiOrZero(s string) int {
	num, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return num
}

// SaveLayout formats the file path using placeholders and metadata from the track and album.
func SaveLayout(props SaveLayoutProps) string {
	logger.Debug("Starting save layout formatting for path: %s", props.Path)

	// Ensure Track and Album are not nil
	if props.Track == nil {
		props.Track = make(map[string]interface{})
	}
	if props.Album == nil {
		props.Album = make(map[string]interface{})
	}

	// Clone album info to avoid modifying the original map
	albumInfo := make(map[string]interface{})
	for k, v := range props.Album {
		albumInfo[k] = v
	}

	// Adjust ALB_TITLE if necessary
	trackDiskNumber, okTrackDisk := props.Track["DISK_NUMBER"]
	albumNumberDisk, okAlbumDisk := props.Album["NUMBER_DISK"]
	albumAlbTitle, okAlbumTitle := albumInfo["ALB_TITLE"]

	if okTrackDisk && okAlbumDisk && okAlbumTitle {
		numDisks := atoiOrZero(fmt.Sprintf("%v", albumNumberDisk))
		if numDisks > 1 {
			albumTitleStr := fmt.Sprintf("%v", albumAlbTitle)
			if !strings.Contains(albumTitleStr, "Disc") {
				discNumber := atoiOrZero(fmt.Sprintf("%v", trackDiskNumber))
				albumInfo["ALB_TITLE"] = fmt.Sprintf("%s (Disc %02d)", albumTitleStr, discNumber)
			}
		}
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

		var value interface{}
		if val, ok := GetNestedValue(albumInfo, key); ok {
			value = val
			logger.Debug("Found value from album: %s = %v", key, value)
		} else if val, ok := GetNestedValue(props.Track, key); ok {
			value = val
			logger.Debug("Found value from track: %s = %v", key, value)
		} else {
			value = ""
		}

		if key == "TRACK_NUMBER" || key == "TRACK_POSITION" || key == "NO_TRACK_NUMBER" {
			if value != "" {
				num := atoiOrZero(fmt.Sprintf("%v", value))
				formattedNum := fmt.Sprintf("%0*d", props.MinimumIntegerDigits, num)
				props.Path = strings.ReplaceAll(props.Path, "{"+key+"}", formattedNum)
				logger.Debug("Formatted track number for key %s: %s", key, props.Path)
			} else {
				props.Path = strings.ReplaceAll(props.Path, "{"+key+"}", "")
				logger.Debug("Key %s had no value; replaced with empty string.", key)
			}
			props.TrackNumber = false
		} else {
			sanitizedValue := SanitizeFileName(fmt.Sprintf("%v", value))
			props.Path = strings.ReplaceAll(props.Path, "{"+key+"}", sanitizedValue)
			logger.Debug("Replaced key %s with sanitized value: %s", key, props.Path)
		}
	}

	if props.TrackNumber {
		var position interface{}
		if pos, exists := props.Track["TRACK_POSITION"]; exists {
			position = pos
		} else if num, exists := props.Track["TRACK_NUMBER"]; exists {
			position = num
		}
		if position != nil {
			num := atoiOrZero(fmt.Sprintf("%v", position))
			trackNumber := fmt.Sprintf("%0*d", props.MinimumIntegerDigits, num)
			dir := filepath.Dir(props.Path)
			base := filepath.Base(props.Path)
			props.Path = filepath.Join(dir, trackNumber+" - "+base)
			logger.Debug("Appended track number to path: %s", props.Path)
		} else {
			props.Path = filepath.Join(props.Path)
		}
	} else {
		props.Path = filepath.Join(props.Path)
	}

	// Remove any remaining problematic characters
	finalPath := strings.Trim(regexp.MustCompile(`[?%*|"<>]`).ReplaceAllString(props.Path, ""), " ")
	logger.Debug("Final sanitized path: %s", finalPath)
	return finalPath
}
