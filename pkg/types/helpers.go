package types

import (
	"encoding/json"
	"fmt"
)

// StringOrInt is a custom type that handles both string and int values during JSON unmarshaling.
type StringOrInt struct {
	Value string
}

// UnmarshalJSON implements the json.Unmarshaler interface for StringOrInt.
func (s *StringOrInt) UnmarshalJSON(data []byte) error {
	// Attempt to unmarshal as a string
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		s.Value = str
		return nil
	}

	// Attempt to unmarshal as an integer
	var num int
	if err := json.Unmarshal(data, &num); err == nil {
		s.Value = fmt.Sprintf("%d", num)
		return nil
	}

	return fmt.Errorf("StringOrInt: failed to unmarshal as string or int")
}
