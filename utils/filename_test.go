package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeFileName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Basic cases
		{"normal_filename.txt", "normal_filename.txt"},
		{"filename with spaces.txt", "filename with spaces.txt"},
		// Invalid characters
		{"file*name?.txt", "file_name_.txt"},
		{"<>:\"/\\|?*\x00-\x1F", "__________-_"},
		// Trailing periods and spaces
		{"filename.", "filename"},
		{"filename   ", "filename"},
		{"filename.  ", "filename"},
		// Combination
		{"inva?lid:fi*le|name.txt", "inva_lid_fi_le_name.txt"},
		// Unicode characters
		{"ファイル名.txt", "ファイル名.txt"},
		// Edge cases
		{"", ""},
		{".", ""},
		{"   ", ""},
		{"...", ""},
		{"..filename..", "..filename"},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := SanitizeFileName(test.input)
			assert.Equal(t, test.expected, result, "SanitizeFileName(%q)", test.input)
		})
	}
}
