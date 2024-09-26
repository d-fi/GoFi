package request

import "encoding/json"

type APIResponse struct {
	Error   []string        `json:"error"`
	Results json.RawMessage `json:"results"`
}

type UserData struct {
	Error   []interface{} `json:"error"`
	Results struct {
		SessionID string `json:"SESSION_ID"`
		Session   string `json:"SESSION"`
		CheckForm string `json:"checkForm"`
	} `json:"results"`
}
