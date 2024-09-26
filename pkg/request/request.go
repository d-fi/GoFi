package request

import (
	"encoding/json"
	"fmt"

	"github.com/d-fi/GoFi/pkg/utils"
)

func checkResponse(data []byte) (json.RawMessage, error) {
	var apiResponse APIResponse
	if err := json.Unmarshal(data, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal API response: %v", err)
	}

	if len(apiResponse.Error) > 0 {
		return nil, fmt.Errorf("API error: %v", apiResponse.Error)
	}

	return apiResponse.Results, nil
}

// Make POST requests to Deezer API
func Request(body map[string]interface{}, method string) ([]byte, error) {
	cacheKey := method + ":" + fmt.Sprintf("%v", body)
	if cachedData, err := getCache(cacheKey); err == nil && len(cachedData) > 0 {
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

	_ = setCache(cacheKey, results)
	return results, nil
}

// Make GET requests to Deezer public API
func RequestGet(method string, params map[string]interface{}, key string) ([]byte, error) {
	cacheKey := method + ":" + key
	if cachedData, err := getCache(cacheKey); err == nil && len(cachedData) > 0 {
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

	_ = setCache(cacheKey, results)
	return results, nil
}

// Make GET requests to Deezer public API
func RequestPublicApi(slug string) ([]byte, error) {
	if cachedData, err := getCache(slug); err == nil && len(cachedData) > 0 {
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
	_ = setCache(slug, results)
	return results, nil
}
