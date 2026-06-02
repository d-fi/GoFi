package download

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

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
var userDataMu sync.Mutex

type deezerUserDataResponse struct {
	Results struct {
		Country string `json:"COUNTRY"`
		User    struct {
			Options struct {
				LicenseToken  string     `json:"license_token"`
				WebLossless   deezerBool `json:"web_lossless"`
				MobileLosless deezerBool `json:"mobile_loseless"`
				WebHQ         deezerBool `json:"web_hq"`
				MobileHQ      deezerBool `json:"mobile_hq"`
			} `json:"OPTIONS"`
		} `json:"USER"`
	} `json:"results"`
}

type deezerBool bool

func (b *deezerBool) UnmarshalJSON(data []byte) error {
	var value any
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	switch value := value.(type) {
	case bool:
		*b = deezerBool(value)
	case string:
		*b = deezerBool(value == "true" || value == "1")
	case float64:
		*b = deezerBool(value != 0)
	default:
		*b = false
	}
	return nil
}

// DzAuthenticate authenticates with Deezer and retrieves user data.
func DzAuthenticate(ctx context.Context) (*UserData, error) {
	userDataMu.Lock()
	defer userDataMu.Unlock()
	if userData != nil {
		logger.Debug("Using cached user data.")
		return userData, nil
	}

	logger.Debug("Authenticating with Deezer to retrieve user data.")
	resp, err := request.Client.R().
		SetContext(ctx).
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

	parsed, err := parseDeezerUserData(resp.Body())
	if err != nil {
		logger.Debug("Failed to parse Deezer user data response: %v", err)
		return nil, err
	}

	userData = parsed
	logger.Debug("Deezer authentication successful. User country: %s", userData.Country)

	return userData, nil
}

func parseDeezerUserData(body []byte) (*UserData, error) {
	var data deezerUserDataResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}
	options := data.Results.User.Options
	if options.LicenseToken == "" {
		return nil, fmt.Errorf("invalid Deezer user data response: missing license token")
	}
	return &UserData{
		LicenseToken:      options.LicenseToken,
		CanStreamLossless: bool(options.WebLossless) || bool(options.MobileLosless),
		CanStreamHQ:       bool(options.WebHQ) || bool(options.MobileHQ),
		Country:           data.Results.Country,
	}, nil
}

// GetTrackUrlFromServer fetches the track URL from the server based on the track token and format.
func GetTrackUrlFromServer(ctx context.Context, trackToken, format string) (string, error) {
	logger.Debug("Fetching track URL from server for format: %s", format)
	user, err := DzAuthenticate(ctx)
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
		SetContext(ctx).
		SetBody(map[string]any{
			"license_token": user.LicenseToken,
			"media": []map[string]any{
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

	var response map[string]any
	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		logger.Debug("Failed to parse track URL response: %v", err)
		return "", err
	}

	data := response["data"].([]any)
	if len(data) > 0 {
		trackData := data[0].(map[string]any)
		if errors, exists := trackData["errors"]; exists {
			errorCode := errors.([]any)[0].(map[string]any)["code"].(float64)
			if errorCode == 2002 {
				logger.Debug("Track is geo-blocked in user's country: %s", user.Country)
				return "", &GeoBlocked{Country: user.Country}
			}
			logger.Debug("API returned an error: %v", errors)
			return "", fmt.Errorf("API error: %v", errors)
		}

		if media := trackData["media"].([]any); len(media) > 0 {
			sources := media[0].(map[string]any)["sources"].([]any)
			trackURL := sources[0].(map[string]any)["url"].(string)
			logger.Debug("Track URL fetched successfully: %s", trackURL)
			return trackURL, nil
		}
	}

	logger.Debug("No valid track URL found in the response.")
	return "", nil
}

// GetTrackDownloadUrl retrieves the download URL of a track based on quality.
func GetTrackDownloadUrl(ctx context.Context, track types.TrackType, quality int) (*TrackDownloadUrl, error) {
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
	url, err := GetTrackUrlFromServer(ctx, track.TRACK_TOKEN, formatName)
	if err == nil && url != "" {
		fileSize, err := utils.CheckURLFileSize(ctx, url, nil)
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
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	logger.Debug("Falling back to old method for track URL.")
	filename := decrypt.GetSongFileName(&decrypt.TrackType{
		MD5_ORIGIN:    track.MD5_ORIGIN,
		SNG_ID:        track.SNG_ID,
		MEDIA_VERSION: track.MEDIA_VERSION,
	}, quality)
	fallbackURL := fmt.Sprintf("https://e-cdns-proxy-%s.dzcdn.net/mobile/1/%s", string(track.MD5_ORIGIN[0]), filename)
	fileSize, err := utils.CheckURLFileSize(ctx, fallbackURL, nil)
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
