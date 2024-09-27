package request

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"time"

	"github.com/d-fi/GoFi/logger"
	"github.com/go-resty/resty/v2"
)

var (
	Client        *resty.Client
	sessionID     string
	refreshTicker *time.Ticker
	userArl       string
)

func init() {
	Client = resty.New().
		SetBaseURL("https://api.deezer.com/1.0").
		SetHeader("Accept", "*/*").
		SetHeader("Accept-Encoding", "gzip, deflate").
		SetHeader("Accept-Language", "en-US").
		SetHeader("Cache-Control", "no-cache").
		SetHeader("Content-Type", "application/json; charset=UTF-8").
		SetHeader("User-Agent", "Deezer/8.32.0.2 (iOS; 14.4; Mobile; en; iPhone10_5)").
		SetQueryParam("version", "8.32.0").
		SetQueryParam("api_key", "ZAIVAHCEISOHWAICUQUEXAEPICENGUAFAEZAIPHAELEEVAHPHUCUFONGUAPASUAY").
		SetQueryParam("output", "3").
		SetQueryParam("input", "3").
		SetQueryParam("buildId", "ios12_universal").
		SetQueryParam("screenHeight", "480").
		SetQueryParam("screenWidth", "320").
		SetQueryParam("lang", "en").
		SetTLSClientConfig(&tls.Config{InsecureSkipVerify: false}).
		SetRetryCount(2).
		SetRetryWaitTime(2 * time.Second).
		SetRetryMaxWaitTime(5 * time.Second)
}

// InitDeezerAPI initializes the Deezer API and sets up a session refresh ticker
func InitDeezerAPI(arl string) (string, error) {
	userArl = arl
	logger.Debug("Initializing Deezer API with ARL length: %d", len(arl))

	if len(arl) != 192 {
		logger.Debug("Invalid ARL length: %d", len(arl))
		return "", fmt.Errorf("invalid arl, length should be 192 characters; you have provided %d characters", len(arl))
	}

	resp, err := Client.R().
		SetHeader("Cookie", "arl="+arl).
		SetQueryParam("method", "deezer.ping").
		SetQueryParam("api_version", "1.0").
		SetQueryParam("api_token", "").
		Get("https://www.deezer.com/ajax/gw-light.php")

	if err != nil {
		logger.Debug("Failed to initialize Deezer API: %v", err)
		return "", fmt.Errorf("failed to initialize Deezer API: %v", err)
	}

	if resp.IsError() {
		logger.Debug("Received error response from Deezer: %v", resp.Status())
		return "", fmt.Errorf("received error response from Deezer: %v", resp.Status())
	}

	var data UserData
	if err := json.Unmarshal(resp.Body(), &data); err != nil {
		logger.Debug("Failed to parse Deezer API response: %v", err)
		return "", fmt.Errorf("failed to parse Deezer API response: %v", err)
	}

	if data.Results.Session == "" {
		logger.Debug("Failed to retrieve session from API response")
		return "", fmt.Errorf("failed to retrieve session from API response")
	}

	sessionID = data.Results.Session
	Client.SetQueryParam("sid", sessionID)
	logger.Debug("Deezer API initialized successfully, session ID: %s", sessionID)

	// Start the session refresh ticker if not already running
	if refreshTicker == nil {
		refreshTicker = time.NewTicker(1 * time.Hour)
		go func() {
			for range refreshTicker.C {
				logger.Debug("Refreshing session ID...")
				_, err := refreshSession()
				if err != nil {
					logger.Error("Failed to refresh session: %v", err)
				} else {
					logger.Debug("Session refreshed successfully.")
				}
			}
		}()
	}

	return sessionID, nil
}

// refreshSession refreshes the Deezer session using the ARL
func refreshSession() (string, error) {
	logger.Debug("Refreshing Deezer session with ARL")

	resp, err := Client.R().
		SetHeader("Cookie", "arl="+userArl).
		SetQueryParam("method", "deezer.ping").
		SetQueryParam("api_version", "1.0").
		SetQueryParam("api_token", "").
		Get("https://www.deezer.com/ajax/gw-light.php")

	if err != nil {
		return "", fmt.Errorf("failed to refresh Deezer session: %v", err)
	}

	if resp.IsError() {
		return "", fmt.Errorf("received error response while refreshing session: %v", resp.Status())
	}

	var data UserData
	if err := json.Unmarshal(resp.Body(), &data); err != nil {
		return "", fmt.Errorf("failed to parse refreshed session response: %v", err)
	}

	if data.Results.Session == "" {
		return "", fmt.Errorf("failed to retrieve refreshed session from response")
	}

	sessionID = data.Results.Session
	Client.SetQueryParam("sid", sessionID)
	logger.Debug("Session refreshed successfully, new session ID: %s", sessionID)
	return sessionID, nil
}
