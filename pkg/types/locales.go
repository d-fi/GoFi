package types

import (
	"encoding/json"
	"fmt"
)

// LocalesType represents locales with their names, where the key is the locale code.
type LocalesType map[string]struct {
	Name string `json:"name"` // The localized name, e.g., 'Daft Punk', 'ダフトパンク' in different languages
}

// UnmarshalJSON for LocalesType allows dynamic handling of multiple structures: objects, empty arrays, etc.
func (lt *LocalesType) UnmarshalJSON(data []byte) error {
	// Attempt to unmarshal directly into the expected map structure
	var mapData map[string]struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(data, &mapData); err == nil {
		*lt = LocalesType(mapData)
		return nil
	}

	// Handle empty array or empty object
	if string(data) == "[]" || string(data) == "{}" {
		*lt = LocalesType{}
		return nil
	}

	// If parsing fails, return an error
	return fmt.Errorf("failed to unmarshal LocalesType: %s", string(data))
}
