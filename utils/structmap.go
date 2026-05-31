package utils

import "encoding/json"

func StructMap(value any) map[string]any {
	if value == nil {
		return map[string]any{}
	}
	if data, ok := value.(map[string]any); ok {
		return data
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return map[string]any{}
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return map[string]any{}
	}
	return out
}
