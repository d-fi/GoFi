package converter

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/d-fi/GoFi/api"
	"github.com/d-fi/GoFi/types"
)

// YouTubeTrackToDeezer converts a YouTube video id to the best matching Deezer track.
func YouTubeTrackToDeezer(id string) (types.TrackType, error) {
	var result types.TrackType
	title, artist, err := fetchYouTubeMetadata(id)
	if err != nil {
		return result, err
	}

	if title != "" && artist != "" {
		search, err := api.SearchAlternative(artist, title, 1)
		if err == nil && len(search.TRACK.Data) > 0 {
			return search.TRACK.Data[0], nil
		}
	}

	query := sanitizeYouTubeTitle(title)
	if query == "" {
		query = sanitizeYouTubeTitle(artist)
	}
	if query != "" {
		search, err := api.SearchMusic(query, 20, "TRACK")
		if err == nil {
			queryLower := strings.ToLower(query)
			for _, track := range search.TRACK.Data {
				if strings.Contains(queryLower, strings.ToLower(track.ART_NAME)) {
					return track, nil
				}
			}
			if len(search.TRACK.Data) > 0 {
				return search.TRACK.Data[0], nil
			}
		}
	}

	return result, fmt.Errorf("no track found for youtube video %s", id)
}

func fetchYouTubeMetadata(id string) (title, artist string, err error) {
	title, artist = fetchYouTubeOEmbed(id)
	if title != "" || artist != "" {
		return title, artist, nil
	}

	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest(http.MethodGet, "https://www.youtube.com/watch?v="+id+"&hl=en", nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}
	body := string(bodyBytes)

	playerJSON, ok := extractJSONAssignment(body, "ytInitialPlayerResponse")
	if ok {
		var player map[string]interface{}
		if err := json.Unmarshal([]byte(playerJSON), &player); err == nil {
			title = stringAt(player, "videoDetails", "title")
			artist = stringAt(player, "videoDetails", "author")
		}
	}
	if title == "" {
		title = extractMetaContent(body, "title")
	}

	initialJSON, ok := extractJSONAssignment(body, "ytInitialData")
	if ok {
		var initial interface{}
		if err := json.Unmarshal([]byte(initialJSON), &initial); err == nil {
			if song := findMetadataRow(initial, "Song"); song != "" {
				title = song
			}
			if rowArtist := findMetadataRow(initial, "Artist"); rowArtist != "" {
				artist = rowArtist
			}
		}
	}

	title = html.UnescapeString(title)
	artist = html.UnescapeString(artist)
	if title == "" && artist == "" {
		return "", "", fmt.Errorf("no track found for youtube video %s", id)
	}
	return title, artist, nil
}

func fetchYouTubeOEmbed(id string) (title, artist string) {
	oembedURL := "https://www.youtube.com/oembed?format=json&url=https://www.youtube.com/watch?v=" + id
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(oembedURL)
	if err != nil {
		return "", ""
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", ""
	}

	var data struct {
		Title      string `json:"title"`
		AuthorName string `json:"author_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", ""
	}
	return html.UnescapeString(data.Title), html.UnescapeString(data.AuthorName)
}

func extractJSONAssignment(body, name string) (string, bool) {
	markers := []string{
		"var " + name + " =",
		"window[\"" + name + "\"] =",
		"window['" + name + "'] =",
		name + " =",
	}

	index := -1
	for _, marker := range markers {
		index = strings.Index(body, marker)
		if index != -1 {
			break
		}
	}
	if index == -1 {
		return "", false
	}
	start := strings.Index(body[index:], "{")
	if start == -1 {
		return "", false
	}
	start += index

	depth := 0
	inString := false
	escaped := false
	for i := start; i < len(body); i++ {
		ch := body[i]
		if inString {
			if escaped {
				escaped = false
				continue
			}
			if ch == '\\' {
				escaped = true
				continue
			}
			if ch == '"' {
				inString = false
			}
			continue
		}

		switch ch {
		case '"':
			inString = true
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return body[start : i+1], true
			}
		}
	}

	return "", false
}

func extractMetaContent(body, name string) string {
	re := regexp.MustCompile(`<meta\s+name="` + regexp.QuoteMeta(name) + `"\s+content="([^"]*)"`)
	matches := re.FindStringSubmatch(body)
	if len(matches) == 2 {
		return html.UnescapeString(matches[1])
	}
	return ""
}

func findMetadataRow(value interface{}, title string) string {
	switch typed := value.(type) {
	case map[string]interface{}:
		if row, ok := typed["metadataRowRenderer"].(map[string]interface{}); ok {
			if stringAt(row, "title", "simpleText") == title {
				return metadataRowContent(row)
			}
		}
		for _, child := range typed {
			if found := findMetadataRow(child, title); found != "" {
				return found
			}
		}
	case []interface{}:
		for _, child := range typed {
			if found := findMetadataRow(child, title); found != "" {
				return found
			}
		}
	}
	return ""
}

func metadataRowContent(row map[string]interface{}) string {
	contents, _ := row["contents"].([]interface{})
	for _, content := range contents {
		contentMap, _ := content.(map[string]interface{})
		if text := stringAt(contentMap, "simpleText"); text != "" {
			return text
		}
		runs, _ := contentMap["runs"].([]interface{})
		if len(runs) > 0 {
			run, _ := runs[0].(map[string]interface{})
			if text, _ := run["text"].(string); text != "" {
				return text
			}
		}
	}
	return ""
}

func stringAt(value map[string]interface{}, path ...string) string {
	var current interface{} = value
	for _, key := range path {
		currentMap, ok := current.(map[string]interface{})
		if !ok {
			return ""
		}
		current = currentMap[key]
	}
	result, _ := current.(string)
	return result
}

func sanitizeYouTubeTitle(title string) string {
	title = strings.ToLower(title)
	replacements := []*regexp.Regexp{
		regexp.MustCompile(`(?i)\(off.*?\)`),
		regexp.MustCompile(`(?i)ft\..*`),
		regexp.MustCompile(`[,\-.]`),
		regexp.MustCompile(`\s+`),
	}
	title = replacements[0].ReplaceAllString(title, "")
	title = replacements[1].ReplaceAllString(title, "")
	title = replacements[2].ReplaceAllString(title, "")
	title = replacements[3].ReplaceAllString(title, " ")
	return strings.TrimSpace(title)
}
