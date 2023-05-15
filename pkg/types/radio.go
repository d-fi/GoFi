package types

type RadioType struct {
	RADIO_ID      string   `json:"RADIO_ID"`
	RADIO_PICTURE string   `json:"RADIO_PICTURE"`
	TITLE         string   `json:"TITLE"`
	TAGS          []string `json:"TAGS"`
	TYPE_INTERNAL string   `json:"__TYPE__"`
}
