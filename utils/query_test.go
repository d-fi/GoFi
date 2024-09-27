package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertToQueryParams(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected map[string]string
	}{
		{
			name:     "Empty map",
			input:    map[string]interface{}{},
			expected: map[string]string{},
		},
		{
			name: "Basic types",
			input: map[string]interface{}{
				"string": "value",
				"int":    42,
				"float":  3.14,
				"bool":   true,
			},
			expected: map[string]string{
				"string": "value",
				"int":    "42",
				"float":  "3.14",
				"bool":   "true",
			},
		},
		{
			name: "Nil values",
			input: map[string]interface{}{
				"key1": nil,
				"key2": "value2",
			},
			expected: map[string]string{
				"key2": "value2",
			},
		},
		{
			name: "Complex types",
			input: map[string]interface{}{
				"slice":    []int{1, 2, 3},
				"map":      map[string]int{"a": 1},
				"function": func() {},
			},
			expected: map[string]string{
				"slice":    "[1 2 3]",
				"map":      "map[a:1]",
				"function": "<nil>",
			},
		},
		{
			name: "Mixed types",
			input: map[string]interface{}{
				"number": 123,
				"nil":    nil,
				"empty":  "",
			},
			expected: map[string]string{
				"number": "123",
				"empty":  "",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := ConvertToQueryParams(test.input)
			assert.Equal(t, test.expected, result, "Test '%s' failed", test.name)
		})
	}
}
