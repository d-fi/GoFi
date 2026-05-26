package dfi

import (
	"bufio"
	"strings"
	"testing"
)

func TestPromptQualityRejectsInvalidChoice(t *testing.T) {
	quality, err := promptQuality(bufio.NewReader(strings.NewReader("4\n2\n")))
	if err != nil {
		t.Fatal(err)
	}
	if quality != "320" {
		t.Fatalf("quality = %q, want 320", quality)
	}
}

func TestPromptQualityMapsMenuChoices(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: "1\n", want: "128"},
		{input: "2\n", want: "320"},
		{input: "3\n", want: "flac"},
	}

	for _, tt := range tests {
		t.Run(strings.TrimSpace(tt.input), func(t *testing.T) {
			quality, err := promptQuality(bufio.NewReader(strings.NewReader(tt.input)))
			if err != nil {
				t.Fatal(err)
			}
			if quality != tt.want {
				t.Fatalf("quality = %q, want %q", quality, tt.want)
			}
		})
	}
}
