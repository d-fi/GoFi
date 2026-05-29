package converter

import (
	"testing"

	"github.com/d-fi/GoFi/types"
	"github.com/stretchr/testify/assert"
)

func TestScoreSpotifyDeezerCandidateAcceptsMatchingTrack(t *testing.T) {
	input := spotifyMatchInput{
		title:       "Blinding Lights",
		artists:     []string{"The Weeknd"},
		album:       "After Hours",
		durationSec: 200,
	}
	candidate := trackCandidate("Blinding Lights", "The Weeknd", "After Hours", 200)

	score := scoreSpotifyDeezerCandidate(input, candidate)

	assert.False(t, score.conflict)
	assert.GreaterOrEqual(t, score.total, spotifyMatchMinScore)
	assert.GreaterOrEqual(t, score.title, spotifyMatchMinTitle)
	assert.GreaterOrEqual(t, score.artist, spotifyMatchMinArtist)
	assert.LessOrEqual(t, score.durationDiff, spotifyMatchMaxDuration)
}

func TestScoreSpotifyDeezerCandidateRejectsUnexpectedLiveVersion(t *testing.T) {
	input := spotifyMatchInput{
		title:       "Wonderwall",
		artists:     []string{"Oasis"},
		album:       "(What's The Story) Morning Glory?",
		durationSec: 259,
	}
	candidate := trackCandidate("Wonderwall (Live from Dublin, 16 August '25)", "Oasis", "Wonderwall", 261)

	score := scoreSpotifyDeezerCandidate(input, candidate)

	assert.True(t, score.conflict)
}

func TestScoreSpotifyDeezerCandidateRejectsUnexpectedRemasterVersion(t *testing.T) {
	input := spotifyMatchInput{
		title:       "No Scrubs",
		artists:     []string{"TLC"},
		album:       "Fanmail",
		durationSec: 214,
	}
	candidate := trackCandidate("No Scrubs (Re-Mastered Version)", "TLC", "Fanmail", 214)

	score := scoreSpotifyDeezerCandidate(input, candidate)

	assert.True(t, score.conflict)
}

func TestScoreSpotifyDeezerCandidateRejectsUnexpectedReRecordedVersion(t *testing.T) {
	input := spotifyMatchInput{
		title:       "Fire Burning",
		artists:     []string{"Sean Kingston"},
		album:       "Tomorrow",
		durationSec: 240,
	}
	candidate := trackCandidate("Fire Burning (Re-Recorded)", "Sean Kingston", "Tomorrow", 240)

	score := scoreSpotifyDeezerCandidate(input, candidate)

	assert.True(t, score.conflict)
}

func TestScoreSpotifyDeezerCandidateRejectsUnexpectedYearVersion(t *testing.T) {
	input := spotifyMatchInput{
		title:       "Too Little Too Late",
		artists:     []string{"JoJo"},
		album:       "The High Road",
		durationSec: 221,
	}
	candidate := trackCandidate("Too Little Too Late (2018)", "JoJo", "The High Road", 221)

	score := scoreSpotifyDeezerCandidate(input, candidate)

	assert.True(t, score.conflict)
}

func TestScoreSpotifyDeezerCandidateAcceptsMatchingRemixVersion(t *testing.T) {
	input := spotifyMatchInput{
		title:       "I Took A Pill In Ibiza - Seeb Remix",
		artists:     []string{"Mike Posner", "Seeb"},
		album:       "At Night, Alone.",
		durationSec: 198,
	}
	candidate := trackCandidate("I Took A Pill In Ibiza (Seeb Remix)", "Mike Posner", "At Night, Alone.", 198)

	score := scoreSpotifyDeezerCandidate(input, candidate)

	assert.False(t, score.conflict)
	assert.GreaterOrEqual(t, score.total, spotifyMatchMinScore)
}

func TestScoreSpotifyDeezerCandidateIgnoresWeakAlbumForExactTrack(t *testing.T) {
	input := spotifyMatchInput{
		title:       "Shivers",
		artists:     []string{"Ed Sheeran"},
		album:       "=",
		durationSec: 207,
	}
	candidate := trackCandidate("Shivers", "Ed Sheeran", "Shivers", 208)

	score := scoreSpotifyDeezerCandidate(input, candidate)

	assert.False(t, score.conflict)
	assert.GreaterOrEqual(t, score.total, spotifyMatchMinScore)
}

func TestSpotifyMatchQueriesIncludeCleanFeatureTitle(t *testing.T) {
	input := spotifyMatchInput{
		title:   "One Kiss (with Dua Lipa)",
		artists: []string{"Calvin Harris", "Dua Lipa"},
		album:   "One Kiss",
	}

	queries := spotifyMatchQueries(input)

	assert.Contains(t, queries, "One Kiss Calvin Harris")
}

func TestCleanSpotifySearchTitleStripsFromSoundtrack(t *testing.T) {
	title := cleanSpotifySearchTitle(`All The Stars (with SZA) - From "Black Panther: The Album"`)

	assert.Equal(t, "All The Stars", title)
}

func TestNormalizeArtistIgnoresLeadingThe(t *testing.T) {
	score := similarityScore(normalizeArtistForCompare("Black Eyed Peas"), normalizeArtistForCompare("The Black Eyed Peas"))

	assert.Equal(t, 100, score)
}

func TestNormalizeArtistHandlesStylizedPunctuation(t *testing.T) {
	tests := map[string]string{
		"Ke$ha":         "Kesha",
		"P!nk":          "Pink",
		"Ty Dolla $ign": "Ty Dolla Sign",
	}
	for left, right := range tests {
		score := artistSimilarityScore(left, right)

		assert.Equal(t, 100, score)
	}
}

func TestNormalizeArtistKeepsExactSymbolVariant(t *testing.T) {
	score := artistSimilarityScore("Ke$ha", "Ke$ha")

	assert.Equal(t, 100, score)
}

func TestVersionTagsDetectHyphenatedRemaster(t *testing.T) {
	tags := versionTags("No Scrubs (Re-Mastered Version)")

	assert.True(t, tags["remaster"])
}

func TestSpotifyPartnerTrackToSpotifyTrack(t *testing.T) {
	partnerTrack := spotifyPartnerTrack{
		URI:  "spotify:track:abc123",
		Name: "Track Name",
		AlbumOfTrack: spotifyPartnerAlbum{
			URI:  "spotify:album:alb123",
			Name: "Album Name",
		},
	}
	partnerTrack.Duration.TotalMilliseconds = 123456
	partnerTrack.Artists.Items = append(partnerTrack.Artists.Items, struct {
		URI     string `json:"uri"`
		Profile struct {
			Name string `json:"name"`
		} `json:"profile"`
	}{
		URI: "spotify:artist:art123",
	})
	partnerTrack.Artists.Items[0].Profile.Name = "Artist Name"

	track := spotifyPartnerTrackToSpotifyTrack(partnerTrack)

	assert.Equal(t, "abc123", track.ID)
	assert.Equal(t, "Track Name", track.Name)
	assert.Equal(t, 123456, track.DurationMS)
	assert.Equal(t, "alb123", track.Album.ID)
	assert.Equal(t, "Album Name", track.Album.Name)
	assert.Len(t, track.Artists, 1)
	assert.Equal(t, "art123", track.Artists[0].ID)
	assert.Equal(t, "Artist Name", track.Artists[0].Name)
}

func trackCandidate(title, artist, album string, duration int) types.TrackType {
	candidate := types.TrackType{}
	candidate.SNG_ID = "1"
	candidate.SNG_TITLE = title
	candidate.ART_NAME = artist
	candidate.ALB_TITLE = album
	candidate.DURATION = types.StringOrInt(duration)
	return candidate
}
