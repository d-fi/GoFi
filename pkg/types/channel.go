package types

type PicturesType struct {
	MD5  string `json:"md5"`
	Type string `json:"type"`
}

type DataType struct {
	Type            string         `json:"type"`
	ID              string         `json:"id"`
	Name            string         `json:"name"`
	Title           string         `json:"title"`
	Logo            *string        `json:"logo"`
	Description     *string        `json:"description"`
	Slug            string         `json:"slug"`
	BackgroundColor string         `json:"background_color"`
	Pictures        []PicturesType `json:"pictures"`
	TYPE_INTERNAL   string         `json:"__TYPE__"`
}

type ChannelDataType struct {
	ItemID          string         `json:"item_id"`
	ID              string         `json:"id"`
	Type            string         `json:"type"`
	Data            []DataType     `json:"data"`
	Target          string         `json:"target"`
	Title           string         `json:"title"`
	Pictures        []PicturesType `json:"pictures"`
	Weight          int            `json:"weight"`
	BackgroundColor string         `json:"background_color"`
}

type ChannelSearchType struct {
	Data  []ChannelDataType `json:"data"`
	Count int               `json:"count"`
	Total int               `json:"total"`
}
