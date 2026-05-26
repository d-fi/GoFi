package utils

import "strings"

// GetNestedValue retrieves a value from a nested map given a dot-separated key.
func GetNestedValue(data map[string]any, key string) (any, bool) {
	keys := strings.Split(key, ".")
	var current any = data
	for _, k := range keys {
		if m, ok := current.(map[string]any); ok {
			if val, exists := m[k]; exists {
				current = val
			} else {
				return nil, false
			}
		} else {
			return nil, false
		}
	}
	return current, true
}
