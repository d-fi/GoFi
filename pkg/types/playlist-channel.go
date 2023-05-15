package types

type NativeAdsType struct {
	AdvertisingData struct {
		PageIDAndroid       string `json:"page_id_android"`
		PageIDAndroidTablet string `json:"page_id_android_tablet"`
		PageIDIpad          string `json:"page_id_ipad"`
		PageIDIphone        string `json:"page_id_iphone"`
		PageIDWeb           string `json:"page_id_web"`
	} `json:"advertising_data"`
	Data   interface{} `json:"data"`
	ID     string      `json:"id"`
	ItemID string      `json:"item_id"`
	Type   string      `json:"type"`
	Weight int         `json:"weight"`
}

type itemType string

type PlaylistChannelItemsType struct {
	ItemID      string      `json:"item_id"`
	ID          string      `json:"id"`
	Type        itemType    `json:"type"`
	Data        interface{} `json:"data"`
	Target      string      `json:"target"`
	Title       string      `json:"title"`
	Subtitle    string      `json:"subtitle"`
	Description string      `json:"description"`
	Pictures    []struct {
		MD5  string `json:"md5"`
		Type string `json:"type"`
	} `json:"pictures"`
	Weight           int `json:"weight"`
	LayoutParameters struct {
		CTA struct {
			Type  string `json:"type"`
			Label string `json:"label"`
		} `json:"cta"`
	} `json:"layout_parameters"`
}

type PlaylistChannelSectionsType struct {
	Layout    string                     `json:"layout"`
	SectionID string                     `json:"section_id"`
	Items     []PlaylistChannelItemsType `json:"items"`
	Title     string                     `json:"title"`
	Target    string                     `json:"target"`
	Related   struct {
		Target    string `json:"target"`
		Label     string `json:"label"`
		Mandatory bool   `json:"mandatory"`
	} `json:"related"`
	Alignment    string `json:"alignment"`
	GroupID      string `json:"group_id"`
	HasMoreItems bool   `json:"hasMoreItems"`
}

type PlaylistChannelType struct {
	Version string `json:"version"`
	PageID  string `json:"page_id"`
	GA      struct {
		ScreenName string `json:"screen_name"`
	} `json:"ga"`
	Title      string                        `json:"title"`
	Persistent bool                          `json:"persistent"`
	Sections   []PlaylistChannelSectionsType `json:"sections"`
	Expire     int                           `json:"expire"`
}
