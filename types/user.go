package types

type UserType struct {
	UserID    string  `json:"USER_ID"`
	Email     string  `json:"EMAIL"`
	Firstname string  `json:"FIRSTNAME"`
	Lastname  string  `json:"LASTNAME"`
	Birthday  string  `json:"BIRTHDAY"`
	BlogName  string  `json:"BLOG_NAME"`
	Sex       string  `json:"SEX"`
	Address   *string `json:"ADDRESS,omitempty"`
	City      *string `json:"CITY,omitempty"`
	Zip       *string `json:"ZIP,omitempty"`
	Country   string  `json:"COUNTRY"`
	Lang      string  `json:"LANG"`
	Phone     *string `json:"PHONE,omitempty"`
	Type      string  `json:"__TYPE__"`
}
