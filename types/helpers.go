package types

import (
	"fmt"
	"strconv"
	"strings"
)

type StringOrInt int

func (s *StringOrInt) UnmarshalJSON(data []byte) error {
	dataStr := strings.Trim(string(data), "\"")
	intValue, err := strconv.Atoi(dataStr)
	if err != nil {
		return fmt.Errorf("StringOrInt: cannot parse '%s' as int", dataStr)
	}
	*s = StringOrInt(intValue)
	return nil
}

type StringOrBool bool

func (s *StringOrBool) UnmarshalJSON(data []byte) error {
	dataStr := strings.Trim(string(data), "\"")
	boolValue, err := strconv.ParseBool(dataStr)
	if err != nil {
		return fmt.Errorf("StringOrBool: cannot parse '%s' as bool", dataStr)
	}
	*s = StringOrBool(boolValue)
	return nil
}
