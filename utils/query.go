package utils

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/d-fi/GoFi/logger"
)

// ConvertToQueryParams converts map[string]interface{} to map[string]string
func ConvertToQueryParams(params map[string]any) map[string]string {
	logger.Debug("Converting parameters to query params: %v", params)

	queryParams := make(map[string]string)
	for key, value := range params {
		if value != nil {
			var convertedValue string
			// Check if the value is a function
			valueType := reflect.TypeOf(value)
			if valueType.Kind() == reflect.Func {
				convertedValue = "<nil>"
			} else if valueType.Kind() == reflect.Map || valueType.Kind() == reflect.Slice || valueType.Kind() == reflect.Array || valueType.Kind() == reflect.Struct {
				data, err := json.Marshal(value)
				if err != nil {
					convertedValue = fmt.Sprintf("%v", value)
				} else {
					convertedValue = string(data)
				}
			} else {
				convertedValue = fmt.Sprintf("%v", value)
			}
			queryParams[key] = convertedValue
			logger.Debug("Converted key: %s, value: %s", key, convertedValue)
		} else {
			logger.Debug("Skipping key with nil value: %s", key)
		}
	}

	logger.Debug("Converted query params: %v", queryParams)
	return queryParams
}
