package metadata

import "testing"

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
