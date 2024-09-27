package download

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/d-fi/GoFi/decrypt"
	"github.com/d-fi/GoFi/logger"
	"github.com/d-fi/GoFi/request"
	"github.com/d-fi/GoFi/types"
	"github.com/d-fi/GoFi/utils"
)

// WrongLicense error for when the user's license doesn't allow streaming certain formats.
type WrongLicense struct {
	Format string
}

func (e *WrongLicense) Error() string {
	return fmt.Sprintf("Your account can't stream %s tracks", e.Format)
}

// GeoBlocked error for when the track is not available in the user's country.
type GeoBlocked struct {
	Country string
}

func (e *GeoBlocked) Error() string {
	return fmt.Sprintf("This track is not available in your country (%s)", e.Country)
}

var userData *UserData

// DzAuthenticate authenticates with Deezer and retrieves user data.
func DzAuthenticate() (*UserData, error) {
	if userData != nil {
		logger.Debug("Using cached user data.")
		return userData, nil
	}

	logger.Debug("Authenticating with Deezer to retrieve user data.")
	resp, err := request.Client.R().
		SetQueryParams(map[string]string{
			"method":      "deezer.getUserData",
			"api_version": "1.0",
			"api_token":   "null",
		}).
		Get("https://www.deezer.com/ajax/gw-light.php")

	if err != nil {
		logger.Debug("Failed to authenticate with Deezer: %v", err)
		return nil, err
	}

	var data map[string]interface{}
	if err := json.Unmarshal(resp.Body(), &data); err != nil {
		logger.Debug("Failed to parse Deezer user data response: %v", err)
		return nil, err
	}

	results := data["results"].(map[string]interface{})
	options := results["USER"].(map[string]interface{})["OPTIONS"].(map[string]interface{})
	country := results["COUNTRY"].(string)

	userData = &UserData{
		LicenseToken:      options["license_token"].(string),
		CanStreamLossless: options["web_lossless"].(bool) || options["mobile_loseless"].(bool),
		CanStreamHQ:       options["web_hq"].(bool) || options["mobile_hq"].(bool),
		Country:           country,
	}
	logger.Debug("Deezer authentication successful. User country: %s", userData.Country)

	return userData, nil
}

// GetTrackUrlFromServer fetches the track URL from the server based on the track token and format.
func GetTrackUrlFromServer(trackToken, format string) (string, error) {
	logger.Debug("Fetching track URL from server for format: %s", format)
	user, err := DzAuthenticate()
	if err != nil {
		logger.Debug("Error during Deezer authentication: %v", err)
		return "", err
	}

	// Check if the user license allows streaming the requested format.
	if (format == "FLAC" && !user.CanStreamLossless) || (format == "MP3_320" && !user.CanStreamHQ) {
		logger.Debug("User license does not allow streaming format: %s", format)
		return "", &WrongLicense{Format: format}
	}

	resp, err := request.Client.R().
		SetBody(map[string]interface{}{
			"license_token": user.LicenseToken,
			"media": []map[string]interface{}{
				{
					"type":    "FULL",
					"formats": []map[string]string{{"format": format, "cipher": "BF_CBC_STRIPE"}},
				},
			},
			"track_tokens": []string{trackToken},
		}).
		Post("https://media.deezer.com/v1/get_url")

	if err != nil {
		logger.Debug("Failed to fetch track URL from server: %v", err)
		return "", err
	}

	var response map[string]interface{}
	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		logger.Debug("Failed to parse track URL response: %v", err)
		return "", err
	}

	data := response["data"].([]interface{})
	if len(data) > 0 {
		trackData := data[0].(map[string]interface{})
		if errors, exists := trackData["errors"]; exists {
			errorCode := errors.([]interface{})[0].(map[string]interface{})["code"].(float64)
			if errorCode == 2002 {
				logger.Debug("Track is geo-blocked in user's country: %s", user.Country)
				return "", &GeoBlocked{Country: user.Country}
			}
			logger.Debug("API returned an error: %v", errors)
			return "", fmt.Errorf("API error: %v", errors)
		}

		if media := trackData["media"].([]interface{}); len(media) > 0 {
			sources := media[0].(map[string]interface{})["sources"].([]interface{})
			trackURL := sources[0].(map[string]interface{})["url"].(string)
			logger.Debug("Track URL fetched successfully: %s", trackURL)
			return trackURL, nil
		}
	}

	logger.Debug("No valid track URL found in the response.")
	return "", nil
}

// GetTrackDownloadUrl retrieves the download URL of a track based on quality.
func GetTrackDownloadUrl(track types.TrackType, quality int) (*TrackDownloadUrl, error) {
	var formatName string
	switch quality {
	case 9:
		formatName = "FLAC"
	case 3:
		formatName = "MP3_320"
	case 1:
		formatName = "MP3_128"
	default:
		logger.Debug("Unknown quality specified: %d", quality)
		return nil, fmt.Errorf("unknown quality %d", quality)
	}

	logger.Debug("Attempting to get track download URL for format: %s", formatName)
	var wrongLicense *WrongLicense
	var geoBlocked *GeoBlocked

	// Attempt to get the URL with the official API.
	url, err := GetTrackUrlFromServer(track.TRACK_TOKEN, formatName)
	if err == nil && url != "" {
		fileSize, err := utils.CheckURLFileSize(url, nil)
		if err == nil && fileSize > 0 {
			logger.Debug("Track URL obtained and verified successfully. File size: %d bytes", fileSize)
			return &TrackDownloadUrl{
				TrackUrl:    url,
				IsEncrypted: strings.Contains(url, "/mobile/") || strings.Contains(url, "/media/"),
				FileSize:    fileSize,
			}, nil
		}
		logger.Debug("Failed to verify track URL or file size: %v", err)
	} else if err != nil {
		if wl, ok := err.(*WrongLicense); ok {
			wrongLicense = wl
		} else if gb, ok := err.(*GeoBlocked); ok {
			geoBlocked = gb
		} else {
			logger.Debug("Error while fetching track URL: %v", err)
			return nil, err
		}
	}

	// Fallback to the old method.
	logger.Debug("Falling back to old method for track URL.")
	filename := decrypt.GetSongFileName(&decrypt.TrackType{
		MD5_ORIGIN:    track.MD5_ORIGIN,
		SNG_ID:        track.SNG_ID,
		MEDIA_VERSION: track.MEDIA_VERSION,
	}, quality)
	fallbackURL := fmt.Sprintf("https://e-cdns-proxy-%s.dzcdn.net/mobile/1/%s", string(track.MD5_ORIGIN[0]), filename)
	fileSize, err := utils.CheckURLFileSize(fallbackURL, nil)
	if err == nil && fileSize > 0 {
		logger.Debug("Fallback URL obtained and verified successfully. File size: %d bytes", fileSize)
		return &TrackDownloadUrl{
			TrackUrl:    fallbackURL,
			IsEncrypted: strings.Contains(fallbackURL, "/mobile/") || strings.Contains(fallbackURL, "/media/"),
			FileSize:    fileSize,
		}, nil
	}

	if wrongLicense != nil {
		logger.Debug("Track URL fetch failed due to license restrictions: %v", wrongLicense)
		return nil, wrongLicense
	}
	if geoBlocked != nil {
		logger.Debug("Track URL fetch failed due to geo-blocking: %v", geoBlocked)
		return nil, geoBlocked
	}

	logger.Debug("Failed to obtain track URL: %v", err)
	return nil, err
}
