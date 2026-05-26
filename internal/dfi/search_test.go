package dfi

import (
	"testing"

	"github.com/d-fi/GoFi/types"
)

func TestParseQualityStrict(t *testing.T) {
	tests := []struct {
		value string
		code  int
		label string
	}{
		{value: "", code: 3, label: "320"},
		{value: "128", code: 1, label: "128"},
		{value: "flac", code: 9, label: "flac"},
	}

	for _, tt := range tests {
		t.Run(tt.label, func(t *testing.T) {
			code, _, label, err := ParseQualityStrict(tt.value)
			if err != nil {
				t.Fatal(err)
			}
			if code != tt.code || label != tt.label {
				t.Fatalf("ParseQualityStrict(%q) = %d/%s, want %d/%s", tt.value, code, label, tt.code, tt.label)
			}
		})
	}
}

func TestParseQualityStrictRejectsInvalidQuality(t *testing.T) {
	if _, _, _, err := ParseQualityStrict("4"); err == nil {
		t.Fatal("ParseQualityStrict(\"4\") returned nil error")
	}
}

func TestSelectTracksByIndexes(t *testing.T) {
	tracks := []types.TrackType{
		{SongType: types.SongType{SNG_ID: "1"}},
		{SongType: types.SongType{SNG_ID: "2"}},
		{SongType: types.SongType{SNG_ID: "3"}},
	}
	selected := SelectTracksByIndexes(tracks, []int{2, 0, 2, 9})
	if len(selected) != 2 {
		t.Fatalf("len(selected) = %d, want 2", len(selected))
	}
	if selected[0].SNG_ID != "3" || selected[1].SNG_ID != "1" {
		t.Fatalf("selected ids = %s/%s, want 3/1", selected[0].SNG_ID, selected[1].SNG_ID)
	}
}
