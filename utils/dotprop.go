package utils

import (
	"reflect"
	"strconv"
	"strings"
)

// GetNestedValue retrieves a value from a nested map given a dot-separated key.
func GetNestedValue(data map[string]any, key string) (any, bool) {
	keys := strings.Split(key, ".")
	var current any = data
	for _, k := range keys {
		value := reflect.ValueOf(current)
		if !value.IsValid() {
			return nil, false
		}
		for value.Kind() == reflect.Pointer || value.Kind() == reflect.Interface {
			if value.IsNil() {
				return nil, false
			}
			value = value.Elem()
		}

		switch value.Kind() {
		case reflect.Map:
			if value.Type().Key().Kind() != reflect.String {
				return nil, false
			}
			item := value.MapIndex(reflect.ValueOf(k))
			if !item.IsValid() {
				return nil, false
			}
			current = item.Interface()
		case reflect.Slice, reflect.Array:
			index, err := strconv.Atoi(k)
			if err != nil || index < 0 || index >= value.Len() {
				return nil, false
			}
			current = value.Index(index).Interface()
		default:
			return nil, false
		}
	}
	return current, true
}
