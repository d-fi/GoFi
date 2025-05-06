package request

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/d-fi/GoFi/logger"
	"github.com/d-fi/GoFi/utils"
	"github.com/hashicorp/golang-lru/v2/expirable"
)

const (
	cacheSize = 1000
	cacheTTL  = 60 * time.Minute
)

var cache = expirable.NewLRU[string, []byte](cacheSize, nil, cacheTTL)

func checkResponse(data []byte) (json.RawMessage, error) {
	logger.Debug("Checking API response")
	var apiResponse APIResponse
	if err := json.Unmarshal(data, &apiResponse); err != nil {
		logger.Debug("Failed to unmarshal API response: %v", err)
		return nil, fmt.Errorf("failed to unmarshal API response: %v", err)
	}

	switch errVal := apiResponse.Error.(type) {
	case string:
		if errVal == "NEED_API_AUTH_REQUIRED" || errVal == "NEED_API_AUTH" {
			// Try to re-initialize with ARL if not already done
			arl := os.Getenv("DEEZER_ARL")
			if arl != "" && !IsInitialized() {
				logger.Debug("Auth required, trying to initialize with ARL from environment...")
				_, initErr := InitDeezerAPI(arl)
				if initErr != nil {
					logger.Error("Failed to initialize with ARL from environment: %v", initErr)
				} else {
					logger.Debug("Successfully initialized with ARL, retrying request")
					return nil, fmt.Errorf("API auth initialized, please retry: %s", errVal)
				}
			}
		}
		logger.Debug("API error: %s", errVal)
		return nil, fmt.Errorf("API error: %s", errVal)
	case map[string]interface{}:
		errorMessage := ""
		for key, value := range errVal {
			errorMessage += fmt.Sprintf("%s: %v, ", key, value)
		}
		logger.Debug("API error: %v", errorMessage)
		return nil, fmt.Errorf("API error: %v", errorMessage)
	}

	logger.Debug("API response checked successfully")
	return apiResponse.Results, nil
}

func Request(body map[string]interface{}, method string) ([]byte, error) {
	cacheKey := method + ":" + fmt.Sprintf("%v", body)
	if cachedData, ok := cache.Get(cacheKey); ok && len(cachedData) > 0 {
		logger.Debug("Cache hit for request with method: %s", method)
		return cachedData, nil
	}

	// Ensure ARL cookie is set
	ensureAuth()

	logger.Debug("Making request with method: %s", method)
	resp, err := Client.R().
		SetBody(body).
		SetQueryParam("method", method).
		Post("/gateway.php")

	if err != nil {
		logger.Debug("Failed to make request: %v", err)
		return nil, err
	}

	responseBody := resp.Body()
	results, err := checkResponse(responseBody)
	if err != nil {
		logger.Debug("Error in response: %v", err)
		return nil, err
	}

	cache.Add(cacheKey, results)
	logger.Debug("Request successful, response cached")
	return results, nil
}

func RequestGet(method string, params map[string]interface{}) ([]byte, error) {
	cacheKey := method + ":get_request"
	if cachedData, ok := cache.Get(cacheKey); ok && len(cachedData) > 0 {
		logger.Debug("Cache hit for GET request with method: %s", method)
		return cachedData, nil
	}

	// Ensure ARL cookie is set
	ensureAuth()

	queryParams := utils.ConvertToQueryParams(params)
	logger.Debug("Making GET request with method: %s", method)
	resp, err := Client.R().
		SetQueryParams(queryParams).
		SetQueryParam("method", method).
		Get("/gateway.php")

	if err != nil {
		logger.Debug("Failed to make GET request: %v", err)
		return nil, err
	}

	responseBody := resp.Body()
	results, err := checkResponse(responseBody)
	if err != nil {
		logger.Debug("Error in GET response: %v", err)
		return nil, err
	}

	cache.Add(cacheKey, results)
	logger.Debug("GET request successful, response cached")
	return results, nil
}

func RequestPublicApi(slug string) ([]byte, error) {
	if cachedData, ok := cache.Get(slug); ok && len(cachedData) > 0 {
		logger.Debug("Cache hit for public API request: %s", slug)
		return cachedData, nil
	}

	// Ensure ARL cookie is set for auth-required endpoints
	ensureAuth()

	logger.Debug("Making public API request: %s", slug)
	resp, err := Client.R().Get("https://api.deezer.com" + slug)
	if err != nil {
		logger.Debug("Failed to make public API request: %v", err)
		return nil, err
	}

	results := resp.Body()

	var errorResponse PublicAPIResponseError
	if err := json.Unmarshal(results, &errorResponse); err == nil {
		if errorResponse.Error.Type != "" {
			logger.Debug("API error: %s - %s (Code: %d)", errorResponse.Error.Type, errorResponse.Error.Message, errorResponse.Error.Code)
			return nil, fmt.Errorf("API error: %s - %s (Code: %d)", errorResponse.Error.Type, errorResponse.Error.Message, errorResponse.Error.Code)
		}
	}

	cache.Add(slug, results)
	logger.Debug("Public API request successful, response cached")
	return results, nil
}

// ensureAuth makes sure the API is initialized with an ARL token if available
func ensureAuth() {
	if !IsInitialized() {
		arl := os.Getenv("DEEZER_ARL")
		if arl != "" {
			logger.Debug("Auto-initializing with ARL from environment...")
			_, err := InitDeezerAPI(arl)
			if err != nil {
				logger.Error("Failed to auto-initialize: %v", err)
			} else {
				logger.Debug("Successfully auto-initialized with ARL")
			}
		}
	}
}