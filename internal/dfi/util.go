package dfi

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/d-fi/GoFi/types"
	"github.com/d-fi/GoFi/utils"
)

func formatSecondsReadable(seconds int) string {
	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	}
	minutes := seconds / 60
	remaining := seconds - minutes*60
	return fmt.Sprintf("%02dm %02ds", minutes, remaining)
}

func commonPath(paths []string) string {
	if len(paths) == 0 {
		return ""
	}
	sorted := append([]string(nil), paths...)
	sort.Strings(sorted)
	first := sorted[0]
	last := sorted[len(sorted)-1]

	i := 0
	for i < len(first) && i < len(last) && first[i] == last[i] {
		i++
	}
	return first[:i]
}

func progressBar(total int64, width int) func(int64) string {
	if width <= 0 {
		width = 40
	}
	unit := float64(total) / float64(width)
	return func(value int64) string {
		chars := width
		if total > 0 && value < total {
			chars = int(float64(value) / unit)
		}
		if chars < 0 {
			chars = 0
		}
		if chars > width {
			chars = width
		}
		return strings.Repeat("█", chars) + strings.Repeat("░", width-chars)
	}
}

func StructMap(value any) map[string]any {
	if value == nil {
		return map[string]any{}
	}
	if data, ok := value.(map[string]any); ok {
		return data
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return map[string]any{}
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return map[string]any{}
	}
	return out
}

func SaveLayout(track types.TrackType, info any, path string, trackNumber bool, totalTracks int) string {
	minDigits := 2
	if totalTracks >= 100 {
		minDigits = 3
	}
	if strings.HasPrefix(path, "{") {
		path = "." + string(filepath.Separator) + path
	}
	return utils.SaveLayout(utils.SaveLayoutProps{
		Track:                StructMap(track),
		Album:                StructMap(info),
		Path:                 path,
		MinimumIntegerDigits: minDigits,
		TrackNumber:          trackNumber,
	})
}

func ParseQuality(value any) (quality int, ext string, label string) {
	switch strings.ToLower(fmt.Sprintf("%v", value)) {
	case "1", "128", "mp3_128", "128kbps":
		return 1, ".mp3", "128"
	case "9", "flac":
		return 9, ".flac", "flac"
	default:
		return 3, ".mp3", "320"
	}
}

func CoverSizeForQuality(sizes CoverSizes, label string) int {
	switch label {
	case "128":
		return sizes.MP3_128
	case "flac":
		return sizes.FLAC
	default:
		return sizes.MP3_320
	}
}

func CoverFilePolicy(tracks []types.TrackType, info any, path string, trackNumber bool) map[string]bool {
	totalTracks := len(tracks)
	albumByDir := map[string]string{}
	allowedByDir := map[string]bool{}
	for _, track := range tracks {
		if track.ALB_PICTURE == "" {
			continue
		}
		dir := coverFilePolicyKey(track, info, path, trackNumber, totalTracks)
		if existing, ok := albumByDir[dir]; ok && existing != track.ALB_PICTURE {
			allowedByDir[dir] = false
			continue
		}
		if _, ok := albumByDir[dir]; !ok {
			albumByDir[dir] = track.ALB_PICTURE
			allowedByDir[dir] = true
		}
	}
	return allowedByDir
}

func coverFilePolicyKey(track types.TrackType, info any, path string, trackNumber bool, totalTracks int) string {
	return coverFileDir(SaveLayout(track, info, path, trackNumber, totalTracks), path)
}

func coverFileDir(savePath, layout string) string {
	dir := filepath.Dir(savePath)
	if strings.Contains(layout, "{DISK_FOLDER}") && filepath.Base(dir) != "." {
		return filepath.Dir(dir)
	}
	return dir
}

func AsInt(value any) int {
	switch v := value.(type) {
	case int:
		return v
	case string:
		n, _ := strconv.Atoi(v)
		return n
	default:
		n, _ := strconv.Atoi(fmt.Sprintf("%v", value))
		return n
	}
}

func LooksLikeURL(value string) bool {
	return strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://") || strings.HasPrefix(value, "spotify:")
}

func uniqueDirs(paths []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, path := range paths {
		dir := filepath.Dir(path)
		if seen[dir] {
			continue
		}
		seen[dir] = true
		out = append(out, dir)
	}
	return out
}
