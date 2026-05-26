package dfi

import (
	"testing"

	"github.com/d-fi/GoFi/types"
)

func TestSignaleMessages(t *testing.T) {
	tests := map[string]string{
		info("hello world"):    "ℹ info hello world",
		warn("hello world"):    "⚠ warn hello world",
		pending("hello world"): "● pending hello world",
		success("hello world"): "✔ success hello world",
		failure("hello world"): "✖ error hello world",
		note("hello world"):    "  → hello world",
	}
	for actual, expected := range tests {
		if actual != expected {
			t.Fatalf("%q != %q", actual, expected)
		}
	}
}

func TestFormatSecondsReadable(t *testing.T) {
	if got := formatSecondsReadable(96); got != "01m 36s" {
		t.Fatalf("formatSecondsReadable(96) = %q", got)
	}
}

func TestSaveLayout(t *testing.T) {
	track := types.TrackType{}
	track.SNG_TITLE = "Harder, Better, Faster, Stronger"
	track.ART_NAME = "Daft Punk"
	track.ALB_TITLE = "Discovery"
	track.TRACK_NUMBER = types.StringOrInt(4)

	layout := SaveLayout(track, map[string]any{"ALB_TITLE": "Discovery"}, "{ALB_TITLE}/{ART_NAME}/{SNG_TITLE}", true, 14)
	if layout != "Discovery/Daft Punk/04 - Harder, Better, Faster, Stronger" {
		t.Fatalf("layout = %q", layout)
	}
}

func TestCommonPath(t *testing.T) {
	got := commonPath([]string{"Playlist/Test", "Playlist/Test/Sub"})
	if got != "Playlist/Test" {
		t.Fatalf("commonPath = %q", got)
	}
}
