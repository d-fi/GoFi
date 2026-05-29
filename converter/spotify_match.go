package converter

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"
	"sync"
	"unicode"

	"github.com/d-fi/GoFi/api"
	"github.com/d-fi/GoFi/types"
	"golang.org/x/text/unicode/norm"
)

const (
	spotifyMatchSearchLimit = 10
	spotifyMatchMinScore    = 88
	spotifyMatchMinTitle    = 85
	spotifyMatchMinArtist   = 80
	spotifyMatchMaxDuration = 5
	spotifyMatchMinLead     = 5
)

var spotifyDeezerSearchLimiter = make(chan struct{}, converterConcurrency)

type spotifyMatchInput struct {
	title       string
	artists     []string
	album       string
	durationSec int
}

type spotifyMatchScore struct {
	total         int
	title         int
	artist        int
	primaryArtist int
	album         int
	duration      int
	durationDiff  int
	conflict      bool
}

// SpotifyTrackMetadataToDeezer matches Spotify track metadata to a Deezer track when ISRC is unavailable.
func SpotifyTrackMetadataToDeezer(track SpotifyTrack) (types.TrackType, error) {
	input := spotifyMatchInput{
		title:       track.Name,
		album:       track.Album.Name,
		durationSec: int(math.Round(float64(track.DurationMS) / 1000)),
	}
	for _, artist := range track.Artists {
		if artist.Name != "" {
			input.artists = append(input.artists, artist.Name)
		}
	}
	return spotifyMetadataToDeezer(input)
}

func spotifyMetadataToDeezer(input spotifyMatchInput) (types.TrackType, error) {
	candidates, err := searchSpotifyDeezerCandidates(input)
	if err != nil {
		return types.TrackType{}, err
	}
	if len(candidates) == 0 {
		return types.TrackType{}, fmt.Errorf("no deezer candidates for spotify track %q", input.title)
	}

	var best *types.TrackType
	var bestScore spotifyMatchScore
	secondBest := 0
	for index := range candidates {
		candidate := candidates[index]
		score := scoreSpotifyDeezerCandidate(input, candidate)
		if score.conflict {
			continue
		}
		if best == nil || score.total > bestScore.total {
			if best != nil {
				secondBest = bestScore.total
			}
			best = &candidate
			bestScore = score
			continue
		}
		if score.total > secondBest {
			secondBest = score.total
		}
	}

	if best == nil {
		return types.TrackType{}, fmt.Errorf("no safe deezer match for spotify track %q", input.title)
	}
	if bestScore.total < spotifyMatchMinScore ||
		bestScore.title < spotifyMatchMinTitle ||
		bestScore.artist < spotifyMatchMinArtist ||
		bestScore.durationDiff > spotifyMatchMaxDuration {
		return types.TrackType{}, fmt.Errorf("weak deezer match for spotify track %q: score=%d title=%d artist=%d duration_diff=%ds", input.title, bestScore.total, bestScore.title, bestScore.artist, bestScore.durationDiff)
	}
	if secondBest > 0 && bestScore.total < 95 && bestScore.total-secondBest < spotifyMatchMinLead {
		return types.TrackType{}, fmt.Errorf("ambiguous deezer match for spotify track %q: best=%d second=%d", input.title, bestScore.total, secondBest)
	}

	return *best, nil
}

func searchSpotifyDeezerCandidates(input spotifyMatchInput) ([]types.TrackType, error) {
	queries := spotifyMatchQueries(input)
	results := make(chan struct {
		tracks []types.TrackType
		err    error
	}, len(queries))

	var wg sync.WaitGroup
	for _, query := range queries {
		wg.Go(func() {
			tracks, err := searchDeezerTracks(query, spotifyMatchSearchLimit)
			results <- struct {
				tracks []types.TrackType
				err    error
			}{tracks: tracks, err: err}
		})
	}
	wg.Wait()
	close(results)

	seen := map[string]bool{}
	var candidates []types.TrackType
	var lastErr error

	for result := range results {
		if result.err != nil {
			lastErr = result.err
			continue
		}
		for _, track := range result.tracks {
			if track.SNG_ID == "" || seen[track.SNG_ID] {
				continue
			}
			seen[track.SNG_ID] = true
			candidates = append(candidates, track)
		}
	}
	if len(candidates) == 0 && lastErr != nil {
		return nil, lastErr
	}
	return candidates, nil
}

func spotifyMatchQueries(input spotifyMatchInput) []string {
	primaryArtist := ""
	if len(input.artists) > 0 {
		primaryArtist = input.artists[0]
	}
	allArtists := strings.Join(input.artists, " ")
	cleanTitle := cleanSpotifySearchTitle(input.title)

	raw := []string{
		strings.TrimSpace(input.title + " " + primaryArtist),
		strings.TrimSpace(cleanTitle + " " + primaryArtist),
		strings.TrimSpace(input.title + " " + allArtists),
		strings.TrimSpace(cleanTitle + " " + allArtists),
		strings.TrimSpace(input.title + " " + primaryArtist + " " + input.album),
		strings.TrimSpace(cleanTitle + " " + primaryArtist + " " + input.album),
	}
	seen := map[string]bool{}
	queries := make([]string, 0, len(raw))
	for _, query := range raw {
		key := normalizeForCompare(query)
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		queries = append(queries, query)
	}
	return queries
}

func searchDeezerTracks(query string, limit int) ([]types.TrackType, error) {
	spotifyDeezerSearchLimiter <- struct{}{}
	defer func() {
		<-spotifyDeezerSearchLimiter
	}()

	search, err := api.SearchMusic(query, limit, "TRACK")
	if err != nil {
		return nil, err
	}
	return search.TRACK.Data, nil
}

func scoreSpotifyDeezerCandidate(input spotifyMatchInput, candidate types.TrackType) spotifyMatchScore {
	title := bestTitleScore(input.title, candidate)
	artist := bestArtistScore(input.artists, candidate)
	primaryArtist := primaryArtistScore(input.artists, candidate.ART_NAME)
	album := similarityScore(normalizeForCompare(input.album), normalizeForCompare(candidate.ALB_TITLE))
	durationDiff := abs(input.durationSec - int(candidate.DURATION))
	duration := durationScore(durationDiff)
	conflict := hasVersionConflict(input.title, candidate) || hasFeatureConflict(input, candidate)

	total := int(math.Round(float64(title)*0.40 + float64(artist)*0.30 + float64(album)*0.15 + float64(duration)*0.15))
	if input.album == "" {
		total = int(math.Round(float64(title)*0.45 + float64(artist)*0.35 + float64(duration)*0.20))
	}
	if title >= 95 && artist >= 90 && durationDiff <= spotifyMatchMaxDuration {
		total = max(total, int(math.Round(float64(title)*0.45+float64(artist)*0.35+float64(duration)*0.20)))
	}
	if durationDiff > 20 || title < 80 || artist < 70 {
		conflict = true
	}
	if primaryArtist < spotifyMatchMinArtist {
		conflict = true
	}

	return spotifyMatchScore{
		total:         total,
		title:         title,
		artist:        artist,
		primaryArtist: primaryArtist,
		album:         album,
		duration:      duration,
		durationDiff:  durationDiff,
		conflict:      conflict,
	}
}

func bestTitleScore(source string, candidate types.TrackType) int {
	score := similarityScore(normalizeTitleBase(source), normalizeTitleBase(candidate.SNG_TITLE))
	if candidate.VERSION != nil && *candidate.VERSION != "" {
		score = max(score, similarityScore(normalizeTitleBase(source), normalizeTitleBase(candidate.SNG_TITLE+" "+*candidate.VERSION)))
	}
	return score
}

func bestArtistScore(sourceArtists []string, candidate types.TrackType) int {
	candidateArtists := []string{candidate.ART_NAME}
	for _, artist := range candidate.ARTISTS {
		if artist.ART_NAME != "" {
			candidateArtists = append(candidateArtists, artist.ART_NAME)
		}
	}

	best := 0
	for _, source := range sourceArtists {
		for _, candidateArtist := range candidateArtists {
			best = max(best, artistSimilarityScore(source, candidateArtist))
		}
	}
	return best
}

func primaryArtistScore(sourceArtists []string, candidatePrimaryArtist string) int {
	best := 0
	for _, source := range sourceArtists {
		best = max(best, artistSimilarityScore(source, candidatePrimaryArtist))
	}
	return best
}

func durationScore(diff int) int {
	switch {
	case diff <= 2:
		return 100
	case diff <= 5:
		return 90
	case diff <= 10:
		return 70
	case diff <= 20:
		return 30
	default:
		return 0
	}
}

func hasVersionConflict(source string, candidate types.TrackType) bool {
	sourceTags := versionTags(source)
	candidateTitle := candidate.SNG_TITLE
	if candidate.VERSION != nil {
		candidateTitle += " " + *candidate.VERSION
	}
	candidateTags := versionTags(candidateTitle)

	for _, tag := range []string{"live", "remix", "acoustic", "instrumental", "karaoke", "cover", "tribute", "sped up", "slowed", "remaster", "re-recorded", "year version"} {
		if candidateTags[tag] && !sourceTags[tag] {
			return true
		}
		if sourceTags[tag] && !candidateTags[tag] {
			return true
		}
	}
	return false
}

func hasFeatureConflict(input spotifyMatchInput, candidate types.TrackType) bool {
	sourceFeatures := featureNames(input.title)
	candidateTitle := candidate.SNG_TITLE
	if candidate.VERSION != nil {
		candidateTitle += " " + *candidate.VERSION
	}
	candidateFeatures := featureNames(candidateTitle)
	if len(candidateFeatures) == 0 {
		return false
	}

	sourceArtists := normalizedArtistSet(input.artists)
	for feature := range sourceFeatures {
		sourceArtists[feature] = true
	}
	for feature := range candidateFeatures {
		if !sourceArtists[feature] {
			return true
		}
	}
	return false
}

func normalizeTitleBase(value string) string {
	value = removeFeatureText(value)
	value = removeNonVersionTitleNoise(value)
	value = removeVersionChunks(value)
	return normalizeForCompare(value)
}

func normalizeArtistForCompare(value string) string {
	value = normalizeForCompare(value)
	value = strings.TrimPrefix(value, "the ")
	return value
}

func artistSimilarityScore(a, b string) int {
	best := 0
	for _, left := range artistCompareVariants(a) {
		for _, right := range artistCompareVariants(b) {
			best = max(best, similarityScore(left, right))
		}
	}
	return best
}

func artistCompareVariants(value string) []string {
	raw := normalizeArtistForCompare(value)
	decoded := normalizeArtistForCompare(strings.NewReplacer(
		"$", "s",
		"!", "i",
		"@", "a",
	).Replace(value))
	if raw == "" || raw == decoded {
		return []string{decoded}
	}
	return []string{raw, decoded}
}

func normalizedArtistSet(artists []string) map[string]bool {
	set := map[string]bool{}
	for _, artist := range artists {
		for _, variant := range artistCompareVariants(artist) {
			if variant != "" {
				set[variant] = true
			}
		}
	}
	return set
}

func featureNames(value string) map[string]bool {
	names := map[string]bool{}
	for _, group := range featureGroups(value) {
		for _, name := range splitFeatureGroup(group) {
			if normalized := normalizeArtistForCompare(name); normalized != "" {
				names[normalized] = true
			}
			for _, variant := range artistCompareVariants(name) {
				if variant != "" {
					names[variant] = true
				}
			}
		}
	}
	return names
}

func featureGroups(value string) []string {
	re := regexp.MustCompile(`(?i)(?:\(|\[)?\b(?:featuring|feat|ft)\.?\s+([^)\]]+)`)
	matches := re.FindAllStringSubmatch(value, -1)
	groups := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) > 1 {
			groups = append(groups, match[1])
		}
	}
	return groups
}

func splitFeatureGroup(value string) []string {
	value = regexp.MustCompile(`(?i)\s+remix\b.*$`).ReplaceAllString(value, "")
	value = strings.NewReplacer(" & ", ",", " and ", ",").Replace(value)
	parts := strings.Split(value, ",")
	names := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			names = append(names, part)
		}
	}
	return names
}

func cleanSpotifySearchTitle(value string) string {
	value = removeFeatureText(value)
	value = removeNonVersionTitleNoise(value)
	return strings.TrimSpace(regexp.MustCompile(`\s+`).ReplaceAllString(value, " "))
}

func normalizeForCompare(value string) string {
	value = strings.ToLower(stripDiacritics(value))
	value = strings.ReplaceAll(value, "&", " and ")

	var builder strings.Builder
	lastSpace := true
	for _, r := range value {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			builder.WriteRune(r)
			lastSpace = false
			continue
		}
		if !lastSpace {
			builder.WriteByte(' ')
			lastSpace = true
		}
	}
	return strings.TrimSpace(builder.String())
}

func stripDiacritics(value string) string {
	decomposed := norm.NFD.String(value)
	var builder strings.Builder
	for _, r := range decomposed {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		builder.WriteRune(r)
	}
	return builder.String()
}

func removeFeatureText(value string) string {
	re := regexp.MustCompile(`(?i)\b(feat|featuring|ft)\.?\b.*`)
	return re.ReplaceAllString(value, "")
}

func removeNonVersionTitleNoise(value string) string {
	value = regexp.MustCompile(`(?i)\s+-\s+from\b.*$`).ReplaceAllString(value, "")
	value = regexp.MustCompile(`(?i)\s+-\s+studio recording\b.*$`).ReplaceAllString(value, "")
	re := regexp.MustCompile(`\([^)]*\)|\[[^]]*\]`)
	return re.ReplaceAllStringFunc(value, func(chunk string) string {
		normalized := normalizeForCompare(chunk)
		for _, marker := range []string{"with ", "feat", "featuring", " from ", "soundtrack", "motion picture", "series", "movie"} {
			if strings.Contains(normalized, marker) {
				return " "
			}
		}
		return chunk
	})
}

func removeVersionChunks(value string) string {
	re := regexp.MustCompile(`\([^)]*\)|\[[^]]*\]| - [^-]+$`)
	return re.ReplaceAllStringFunc(value, func(chunk string) string {
		if len(versionTags(chunk)) > 0 {
			return " "
		}
		return chunk
	})
}

func versionTags(value string) map[string]bool {
	normalized := normalizeForCompare(value)
	tags := map[string]bool{}
	checks := map[string][]string{
		"live":         {"live"},
		"remix":        {"remix"},
		"acoustic":     {"acoustic", "unplugged"},
		"instrumental": {"instrumental"},
		"karaoke":      {"karaoke"},
		"cover":        {"cover"},
		"tribute":      {"tribute"},
		"sped up":      {"sped up", "speed up"},
		"slowed":       {"slowed"},
		"remaster":     {"remaster", "remastered", "re master", "re mastered"},
		"re-recorded":  {"re recorded", "rerecorded"},
		"radio edit":   {"radio edit"},
		"extended":     {"extended"},
	}
	for tag, patterns := range checks {
		for _, pattern := range patterns {
			if strings.Contains(normalized, pattern) {
				tags[tag] = true
				break
			}
		}
	}
	if hasYearVersion(value) {
		tags["year version"] = true
	}
	return tags
}

func hasYearVersion(value string) bool {
	re := regexp.MustCompile(`(?i)(\([^)]*\b(?:19|20)\d{2}\b[^)]*\)|\[[^]]*\b(?:19|20)\d{2}\b[^]]*\]|\b(?:19|20)\d{2}\s+version\b|-\s*(?:19|20)\d{2}\s*$)`)
	return re.MatchString(value)
}

func similarityScore(a, b string) int {
	if a == "" || b == "" {
		return 0
	}
	if a == b {
		return 100
	}
	direct := levenshteinSimilarity(a, b)
	token := levenshteinSimilarity(sortTokens(a), sortTokens(b))
	return max(direct, token)
}

func sortTokens(value string) string {
	tokens := strings.Fields(value)
	sort.Strings(tokens)
	return strings.Join(tokens, " ")
}

func levenshteinSimilarity(a, b string) int {
	ar := []rune(a)
	br := []rune(b)
	maxLen := max(len(ar), len(br))
	if maxLen == 0 {
		return 100
	}
	distance := levenshteinDistance(ar, br)
	return int(math.Round((1 - float64(distance)/float64(maxLen)) * 100))
}

func levenshteinDistance(a, b []rune) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}

	prev := make([]int, len(b)+1)
	curr := make([]int, len(b)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		curr[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			curr[j] = min(curr[j-1]+1, prev[j]+1, prev[j-1]+cost)
		}
		prev, curr = curr, prev
	}
	return prev[len(b)]
}

func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
