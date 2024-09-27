package request

import "encoding/json"

type APIResponse struct {
	Error   interface{}     `json:"error"`
	Results json.RawMessage `json:"results"`
	Payload interface{}     `json:"payload,omitempty"`
}

type PublicAPIResponseError struct {
	Error struct {
		Type    string `json:"type"`
		Message string `json:"message"`
		Code    int    `json:"code"`
	} `json:"error"`
}

type UserData struct {
	Error   []interface{} `json:"error"`
	Results struct {
		SessionID string `json:"SESSION_ID"`
		Session   string `json:"SESSION"`
		CheckForm string `json:"checkForm"`
	} `json:"results"`
}
