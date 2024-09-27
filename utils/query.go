package utils

import "fmt"

// ConvertToQueryParams converts map[string]interface{} to map[string]string
func ConvertToQueryParams(params map[string]interface{}) map[string]string {
	queryParams := make(map[string]string)
	for key, value := range params {
		queryParams[key] = fmt.Sprintf("%v", value)
	}
	return queryParams
}
