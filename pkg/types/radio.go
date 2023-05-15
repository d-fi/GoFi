package types

type RadioType struct {
	RADIO_ID      string   `json:"RADIO_ID"`
	RADIO_PICTURE string   `json:"RADIO_PICTURE"`
	TITLE         string   `json:"TITLE"`
	TAGS          []string `json:"TAGS"`
	__TYPE__      string   `json:"__TYPE__"`
}
