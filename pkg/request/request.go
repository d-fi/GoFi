package request

import (
	"fmt"

	"github.com/d-fi/GoFi/pkg/utils"
)

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
	_ = setCache(cacheKey, responseBody)
	return responseBody, nil
}

// Make POST requests to Deezer light API
func RequestLight(body map[string]interface{}, method string) ([]byte, error) {
	cacheKey := method + ":" + fmt.Sprintf("%v", body)
	if cachedData, err := getCache(cacheKey); err == nil && len(cachedData) > 0 {
		return cachedData, nil
	}

	resp, err := Client.R().
		SetBody(body).
		SetQueryParams(map[string]string{
			"method":      method,
			"api_version": "1.0",
		}).
		Post("https://www.deezer.com/ajax/gw-light.php")

	if err != nil {
		return nil, err
	}

	responseBody := resp.Body()
	_ = setCache(cacheKey, responseBody)
	return responseBody, nil
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
	_ = setCache(cacheKey, responseBody)
	return responseBody, nil
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

	responseBody := resp.Body()
	_ = setCache(slug, responseBody)
	return responseBody, nil
}
