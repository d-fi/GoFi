package utils

import (
	"fmt"
	"maps"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/d-fi/GoFi/logger"
)

// SaveLayoutProps holds the parameters required for the SaveLayout function.
type SaveLayoutProps struct {
	Track                map[string]any
	Album                map[string]any
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
		props.Track = make(map[string]any)
	}
	if props.Album == nil {
		props.Album = make(map[string]any)
	}

	// Clone album info to avoid modifying the original map
	albumInfo := make(map[string]any)
	maps.Copy(albumInfo, props.Album)

	usesDiskFolder := strings.Contains(props.Path, "{DISK_FOLDER}")
	if _, ok := albumInfo["DISK_FOLDER"]; !ok && usesDiskFolder {
		if diskFolder := diskFolder(props.Track, props.Album); diskFolder != "" {
			albumInfo["DISK_FOLDER"] = diskFolder
		}
	}

	if !usesDiskFolder {
		adjustAlbumTitleForDisc(props.Track, props.Album, albumInfo)
	}

	if _, ok := albumInfo["RELEASE_DATE"]; !ok {
		for _, key := range ReleaseDateKeys() {
			var value any
			var exists bool
			if value, exists = GetNestedValue(albumInfo, key); !exists {
				value, exists = GetNestedValue(props.Track, key)
			}
			if !exists {
				continue
			}
			date := fmt.Sprintf("%v", value)
			if date == "" || date == "<nil>" || date == "0000-00-00" {
				continue
			}
			albumInfo["RELEASE_DATE"] = date
			if year := ReleaseYear(date); year != "" {
				albumInfo["RELEASE_YEAR"] = year
			}
			break
		}
	}

	// Find keys inside {}
	re := regexp.MustCompile(`\{([^}]*)\}`)
	matches := re.FindAllStringSubmatch(props.Path, -1)

	for _, match := range matches {
		expression := match[1]
		logger.Debug("Processing key: %s", expression)

		key, value := resolveLayoutValue(albumInfo, props.Track, expression)
		if isTrackNumberLayoutKey(expression) || isTrackNumberLayoutKey(key) {
			if !isEmptyLayoutValue(value) {
				num := atoiOrZero(fmt.Sprintf("%v", value))
				formattedNum := fmt.Sprintf("%0*d", props.MinimumIntegerDigits, num)
				props.Path = strings.ReplaceAll(props.Path, "{"+expression+"}", formattedNum)
				logger.Debug("Formatted track number for key %s: %s", expression, props.Path)
			} else {
				props.Path = strings.ReplaceAll(props.Path, "{"+expression+"}", "")
				logger.Debug("Key %s had no value; replaced with empty string.", expression)
			}
			props.TrackNumber = false
		} else {
			sanitizedValue := SanitizeFileName(fmt.Sprintf("%v", value))
			props.Path = strings.ReplaceAll(props.Path, "{"+expression+"}", sanitizedValue)
			logger.Debug("Replaced key %s with sanitized value: %s", expression, props.Path)
		}
	}

	if props.TrackNumber {
		var position any
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

func adjustAlbumTitleForDisc(track, album, albumInfo map[string]any) {
	trackDiskNumber, okTrackDisk := track["DISK_NUMBER"]
	albumNumberDisk, okAlbumDisk := album["NUMBER_DISK"]
	albumAlbTitle, okAlbumTitle := albumInfo["ALB_TITLE"]
	if !okTrackDisk || !okAlbumDisk || !okAlbumTitle {
		return
	}

	numDisks := atoiOrZero(fmt.Sprintf("%v", albumNumberDisk))
	if numDisks <= 1 {
		return
	}
	albumTitleStr := fmt.Sprintf("%v", albumAlbTitle)
	if strings.Contains(albumTitleStr, "Disc") {
		return
	}
	discNumber := atoiOrZero(fmt.Sprintf("%v", trackDiskNumber))
	albumInfo["ALB_TITLE"] = fmt.Sprintf("%s (Disc %02d)", albumTitleStr, discNumber)
}

func diskFolder(track, album map[string]any) string {
	numDisks := valueAsInt(album, "NUMBER_DISK")
	if numDisks <= 1 {
		return ""
	}
	discNumber := valueAsInt(track, "DISK_NUMBER")
	if discNumber <= 0 {
		return ""
	}
	return fmt.Sprintf("CD%d", discNumber)
}

func valueAsInt(data map[string]any, key string) int {
	value, ok := data[key]
	if !ok {
		return 0
	}
	return atoiOrZero(fmt.Sprintf("%v", value))
}

func ReleaseDateKeys() []string {
	return []string{
		"ORIGINAL_RELEASE_DATE",
		"PHYSICAL_RELEASE_DATE",
		"release_date",
		"album.release_date",
		"DIGITAL_RELEASE_DATE",
		"DATE_START",
	}
}

func ReleaseYear(date string) string {
	date = strings.TrimSpace(date)
	if date == "" || date == "0000-00-00" {
		return ""
	}
	if year, _, ok := strings.Cut(date, "-"); ok && len(year) == 4 {
		return year
	}
	parts := strings.Split(date, "/")
	if len(parts) == 3 && len(parts[2]) == 4 {
		return parts[2]
	}
	return ""
}

func resolveLayoutValue(album, track map[string]any, expression string) (string, any) {
	for key := range strings.SplitSeq(expression, "|") {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		if value, ok := GetNestedValue(album, key); ok && !isEmptyLayoutValue(value) {
			logger.Debug("Found value from album: %s = %v", key, value)
			return key, value
		}
		if value, ok := GetNestedValue(track, key); ok && !isEmptyLayoutValue(value) {
			logger.Debug("Found value from track: %s = %v", key, value)
			return key, value
		}
	}
	return expression, ""
}

func isEmptyLayoutValue(value any) bool {
	if value == nil {
		return true
	}
	text := strings.TrimSpace(fmt.Sprintf("%v", value))
	return text == "" || text == "<nil>"
}

func isTrackNumberLayoutKey(key string) bool {
	if key == "TRACK_NUMBER" || key == "TRACK_POSITION" || key == "NO_TRACK_NUMBER" {
		return true
	}
	return strings.HasSuffix(key, "TRACK_NUMBER") || strings.HasSuffix(key, "TRACK_POSITION")
}
