package metadata

import (
	"testing"

	"github.com/d-fi/GoFi/types"
)

func TestCoverModeDefaultsToEmbed(t *testing.T) {
	if got := NormalizeCoverMode(""); got != CoverModeEmbed {
		t.Fatalf("NormalizeCoverMode empty = %q, want embed", got)
	}
	if !ShouldEmbedCover("") {
		t.Fatal("empty cover mode should embed cover")
	}
	if ShouldSaveCoverFile("") {
		t.Fatal("empty cover mode should not save cover file")
	}
}

func TestCoverModeFileDoesNotEmbed(t *testing.T) {
	if ShouldEmbedCover(CoverModeFile) {
		t.Fatal("file cover mode should not embed cover")
	}
	if !ShouldSaveCoverFile(CoverModeFile) {
		t.Fatal("file cover mode should save cover file")
	}
}

func TestNormalizeCoverFileName(t *testing.T) {
	tests := map[string]string{
		"":                 "cover.jpg",
		"folder":           "folder.jpg",
		"folder.jpeg":      "folder.jpeg",
		"../Folder.png":    "Folder.jpg",
		"nested/cover.jpg": "cover.jpg",
	}
	for input, expected := range tests {
		if got := NormalizeCoverFileName(input); got != expected {
			t.Fatalf("NormalizeCoverFileName(%q) = %q, want %q", input, got, expected)
		}
	}
}

func TestTagReleaseDatePrefersPrivateAlbumDate(t *testing.T) {
	publicAlbum := &types.AlbumTypePublicApi{ReleaseDate: "2011-06-10"}
	privateAlbum := map[string]any{
		"ORIGINAL_RELEASE_DATE": "1990-10-29",
	}

	got := tagReleaseDate(publicAlbum, privateAlbum, types.TrackType{})
	if got != "1990-10-29" {
		t.Fatalf("tagReleaseDate = %q, want 1990-10-29", got)
	}
}

func TestTagReleaseDateFallsBackToPublicAlbumDate(t *testing.T) {
	publicAlbum := &types.AlbumTypePublicApi{ReleaseDate: "2011-06-10"}
	track := types.TrackType{SongType: types.SongType{DATE_START: "1990-10-29"}}

	got := tagReleaseDate(publicAlbum, nil, track)
	if got != "2011-06-10" {
		t.Fatalf("tagReleaseDate = %q, want 2011-06-10", got)
	}
}
