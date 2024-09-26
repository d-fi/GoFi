package api

import (
	"encoding/json"
	"fmt"

	"github.com/d-fi/GoFi/pkg/utils"
	"github.com/go-resty/resty/v2"
)

var client = resty.New()

// Make POST requests to Deezer API
func Request(body map[string]interface{}, method string) (map[string]interface{}, error) {
	cacheKey := method + ":" + fmt.Sprintf("%v", body)
	if cachedData, err := getCache(cacheKey); err == nil && len(cachedData) > 0 {
		var results map[string]interface{}
		if err := json.Unmarshal(cachedData, &results); err == nil {
			return results, nil
		}
	}

	resp, err := client.R().
		SetBody(body).
		SetQueryParam("method", method).
		Post("/gateway.php")

	if err != nil {
		return nil, err
	}

	var responseData struct {
		Error   map[string]interface{} `json:"error"`
		Results map[string]interface{} `json:"results"`
	}

	if err := json.Unmarshal(resp.Body(), &responseData); err != nil {
		return nil, err
	}

	if len(responseData.Results) > 0 {
		cacheData, _ := json.Marshal(responseData.Results)
		_ = setCache(cacheKey, cacheData)
		return responseData.Results, nil
	}

	errorMessage := fmt.Sprintf("%v", responseData.Error)
	return nil, fmt.Errorf("API error: %s", errorMessage)
}

// Make POST requests to Deezer light API
func RequestLight(body map[string]interface{}, method string) (map[string]interface{}, error) {
	cacheKey := method + ":" + fmt.Sprintf("%v", body)
	if cachedData, err := getCache(cacheKey); err == nil && len(cachedData) > 0 {
		var results map[string]interface{}
		if err := json.Unmarshal(cachedData, &results); err == nil {
			return results, nil
		}
	}

	resp, err := client.R().
		SetBody(body).
		SetQueryParams(map[string]string{
			"method":      method,
			"api_version": "1.0",
		}).
		Post("https://www.deezer.com/ajax/gw-light.php")

	if err != nil {
		return nil, err
	}

	var responseData struct {
		Error   map[string]interface{} `json:"error"`
		Results map[string]interface{} `json:"results"`
	}

	if err := json.Unmarshal(resp.Body(), &responseData); err != nil {
		return nil, err
	}

	if len(responseData.Results) > 0 {
		cacheData, _ := json.Marshal(responseData.Results)
		_ = setCache(cacheKey, cacheData)
		return responseData.Results, nil
	}

	errorMessage := fmt.Sprintf("%v", responseData.Error)
	return nil, fmt.Errorf("API error: %s", errorMessage)
}

// Make GET requests to Deezer public API
func RequestGet(method string, params map[string]interface{}, key string) (map[string]interface{}, error) {
	cacheKey := method + key
	if cachedData, err := getCache(cacheKey); err == nil && len(cachedData) > 0 {
		var results map[string]interface{}
		if err := json.Unmarshal(cachedData, &results); err == nil {
			return results, nil
		}
	}

	queryParams := utils.ConvertToQueryParams(params)
	resp, err := client.R().
		SetQueryParams(queryParams).
		SetQueryParam("method", method).
		Get("/gateway.php")

	if err != nil {
		return nil, err
	}

	var responseData struct {
		Error   map[string]interface{} `json:"error"`
		Results map[string]interface{} `json:"results"`
	}

	if err := json.Unmarshal(resp.Body(), &responseData); err != nil {
		return nil, err
	}

	if len(responseData.Results) > 0 {
		cacheData, _ := json.Marshal(responseData.Results)
		_ = setCache(cacheKey, cacheData)
		return responseData.Results, nil
	}

	errorMessage := fmt.Sprintf("%v", responseData.Error)
	return nil, fmt.Errorf("API error: %s", errorMessage)
}

// Make GET requests to Deezer public API
func RequestPublicApi(slug string) (map[string]interface{}, error) {
	if cachedData, err := getCache(slug); err == nil && len(cachedData) > 0 {
		var data map[string]interface{}
		if err := json.Unmarshal(cachedData, &data); err == nil {
			return data, nil
		}
	}

	resp, err := client.R().Get("https://api.deezer.com" + slug)
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	if err := json.Unmarshal(resp.Body(), &data); err != nil {
		return nil, err
	}

	if _, exists := data["error"]; exists {
		errorMessage := fmt.Sprintf("%v", data["error"])
		return nil, fmt.Errorf("API error: %s", errorMessage)
	}

	cacheData, _ := json.Marshal(data)
	_ = setCache(slug, cacheData)
	return data, nil
}
