package request

import (
	"encoding/json"
	"fmt"

	"github.com/d-fi/GoFi/pkg/cache"
	"github.com/d-fi/GoFi/pkg/utils"
)

func checkResponse(data []byte) (json.RawMessage, error) {
	var apiResponse APIResponse
	if err := json.Unmarshal(data, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal API response: %v", err)
	}

	// Check if the response contains error data in different formats
	switch errVal := apiResponse.Error.(type) {
	case string:
		return nil, fmt.Errorf("API error: %s", errVal)
	case map[string]interface{}:
		// Convert the map to a string for better error message readability
		errorMessage := ""
		for key, value := range errVal {
			errorMessage += fmt.Sprintf("%s: %v, ", key, value)
		}
		return nil, fmt.Errorf("API error: %v", errorMessage)
	}

	return apiResponse.Results, nil
}

// Make POST requests to Deezer API
func Request(body map[string]interface{}, method string) ([]byte, error) {
	cacheKey := method + ":" + fmt.Sprintf("%v", body)
	if cachedData, err := cache.GetCache(cacheKey); err == nil && len(cachedData) > 0 {
		return cachedData, nil
	}

	resp, err := Client.R().
		SetBody(body).
		SetQueryParam("method", method).
		Post("/gateway.php")

	if err != nil {
		return nil, err
	}

	responseBody := resp.Body()
	results, err := checkResponse(responseBody)
	if err != nil {
		return nil, err
	}

	_ = cache.SetCache(cacheKey, results)
	return results, nil
}

// Make GET requests to Deezer public API
func RequestGet(method string, params map[string]interface{}) ([]byte, error) {
	cacheKey := method + ":get_request"
	if cachedData, err := cache.GetCache(cacheKey); err == nil && len(cachedData) > 0 {
		return cachedData, nil
	}

	queryParams := utils.ConvertToQueryParams(params)
	resp, err := Client.R().
		SetQueryParams(queryParams).
		SetQueryParam("method", method).
		Get("/gateway.php")

	if err != nil {
		return nil, err
	}

	responseBody := resp.Body()
	results, err := checkResponse(responseBody)
	if err != nil {
		return nil, err
	}

	_ = cache.SetCache(cacheKey, results)
	return results, nil
}

// Make GET requests to Deezer public API
func RequestPublicApi(slug string) ([]byte, error) {
	if cachedData, err := cache.GetCache(slug); err == nil && len(cachedData) > 0 {
		return cachedData, nil
	}

	resp, err := Client.R().Get("https://api.deezer.com" + slug)
	if err != nil {
		return nil, err
	}

	results := resp.Body()

	var errorResponse PublicAPIResponseError

	// Unmarshal response to check for errors
	if err := json.Unmarshal(results, &errorResponse); err == nil {
		if errorResponse.Error.Type != "" {
			return nil, fmt.Errorf("API error: %s - %s (Code: %d)", errorResponse.Error.Type, errorResponse.Error.Message, errorResponse.Error.Code)
		}
	}

	// Cache the response if there are no errors
	_ = cache.SetCache(slug, results)
	return results, nil
}
