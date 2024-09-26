package request

import (
	"encoding/json"
	"fmt"

	"github.com/go-resty/resty/v2"
)

var Client = resty.New().SetBaseURL("https://api.deezer.com/1.0").
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
	SetQueryParam("lang", "en")

func InitDeezerAPI(arl string) (string, error) {
	if len(arl) != 192 {
		return "", fmt.Errorf("Invalid arl. Length should be 192 characters. You have provided %d characters.", len(arl))
	}

	resp, err := Client.R().
		SetHeader("Cookie", "arl="+arl).
		SetQueryParam("method", "deezer.ping").
		SetQueryParam("api_version", "1.0").
		SetQueryParam("api_token", "").
		Get("https://www.deezer.com/ajax/gw-light.php")
	if err != nil {
		return "", err
	}

	var data UserData
	if err := json.Unmarshal(resp.Body(), &data); err != nil {
		return "", err
	}

	Client.SetQueryParam("sid", data.Results.Session)
	return data.Results.Session, nil
}
