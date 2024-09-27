package types

type ProfileTypeMinimal struct {
	USER_ID       string `json:"USER_ID"`
	FIRSTNAME     string `json:"FIRSTNAME"`
	LASTNAME      string `json:"LASTNAME"`
	BLOG_NAME     string `json:"BLOG_NAME"`
	USER_PICTURE  string `json:"USER_PICTURE,omitempty"`
	IS_FOLLOW     bool   `json:"IS_FOLLOW"`
	TYPE_INTERNAL string `json:"__TYPE__"`
}

type ProfileType struct {
	IS_FOLLOW     bool            `json:"IS_FOLLOW"`
	NB_ARTISTS    int             `json:"NB_ARTISTS"`
	NB_FOLLOWERS  int             `json:"NB_FOLLOWERS"`
	NB_FOLLOWINGS int             `json:"NB_FOLLOWINGS"`
	NB_MP3S       int             `json:"NB_MP3S"`
	TOP_TRACK     AlbumTracksType `json:"TOP_TRACK"`
	USER          UserProfileType `json:"USER"`
}

type UserProfileType struct {
	USER_ID       string `json:"USER_ID"`
	BLOG_NAME     string `json:"BLOG_NAME"`
	SEX           string `json:"SEX,omitempty"`
	COUNTRY       string `json:"COUNTRY"`
	USER_PICTURE  string `json:"USER_PICTURE,omitempty"`
	COUNTRY_NAME  string `json:"COUNTRY_NAME"`
	PRIVATE       bool   `json:"PRIVATE"`
	DISPLAY_NAME  string `json:"DISPLAY_NAME"`
	TYPE_INTERNAL string `json:"__TYPE__"`
}
