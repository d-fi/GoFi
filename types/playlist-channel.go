package types

import "encoding/json"

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

type PlaylistChannelItemType string

const (
	PlaylistChannelItemAlbum      PlaylistChannelItemType = "album"
	PlaylistChannelItemArtist     PlaylistChannelItemType = "artist"
	PlaylistChannelItemChannel    PlaylistChannelItemType = "channel"
	PlaylistChannelItemEpisode    PlaylistChannelItemType = "episode"
	PlaylistChannelItemLivestream PlaylistChannelItemType = "livestream"
	PlaylistChannelItemNative     PlaylistChannelItemType = "native"
	PlaylistChannelItemPage       PlaylistChannelItemType = "page"
	PlaylistChannelItemPlaylist   PlaylistChannelItemType = "playlist"
	PlaylistChannelItemRadio      PlaylistChannelItemType = "radio"
	PlaylistChannelItemShow       PlaylistChannelItemType = "show"
	PlaylistChannelItemTrack      PlaylistChannelItemType = "track"
)

type PlaylistChannelPicture struct {
	MD5  string `json:"md5"`
	Type string `json:"type"`
}

type PlaylistChannelPlaylistData struct {
	DateAdd         string      `json:"DATE_ADD"`
	DateMod         string      `json:"DATE_MOD"`
	Description     string      `json:"DESCRIPTION"`
	NbFan           int         `json:"NB_FAN"`
	NbSong          int         `json:"NB_SONG"`
	ParentUserID    string      `json:"PARENT_USER_ID"`
	ParentUsername  string      `json:"PARENT_USERNAME,omitempty"`
	PictureType     string      `json:"PICTURE_TYPE"`
	PlaylistID      string      `json:"PLAYLIST_ID"`
	PlaylistPicture string      `json:"PLAYLIST_PICTURE"`
	Status          StringOrInt `json:"STATUS"`
	Title           string      `json:"TITLE"`
	Type            string      `json:"TYPE"`
	TypeInternal    string      `json:"__TYPE__"`
}

type PlaylistChannelAlbumData struct {
	AlbumID              string       `json:"ALB_ID"`
	AlbumPicture         string       `json:"ALB_PICTURE"`
	AlbumTitle           string       `json:"ALB_TITLE"`
	ArtistID             string       `json:"ART_ID"`
	ArtistName           string       `json:"ART_NAME"`
	Artists              []ArtistType `json:"ARTISTS"`
	Available            bool         `json:"AVAILABLE"`
	DigitalReleaseDate   string       `json:"DIGITAL_RELEASE_DATE"`
	ExplicitLyrics       string       `json:"EXPLICIT_LYRICS"`
	NbFan                int          `json:"NB_FAN"`
	NumberDisk           string       `json:"NUMBER_DISK"`
	NumberTrack          string       `json:"NUMBER_TRACK"`
	PhysicalReleaseDate  string       `json:"PHYSICAL_RELEASE_DATE"`
	ProducerLine         string       `json:"PRODUCER_LINE"`
	Rank                 string       `json:"RANK"`
	Status               string       `json:"STATUS"`
	Type                 string       `json:"TYPE"`
	TypeInternal         string       `json:"__TYPE__"`
	UPC                  string       `json:"UPC"`
	Version              string       `json:"VERSION"`
	ExplicitAlbumContent struct {
		ExplicitLyricsStatus int `json:"EXPLICIT_LYRICS_STATUS"`
		ExplicitCoverStatus  int `json:"EXPLICIT_COVER_STATUS"`
	} `json:"EXPLICIT_ALBUM_CONTENT"`
}

type PlaylistChannelLivestreamData struct {
	BackgroundColor string                   `json:"background_color"`
	Description     *string                  `json:"description"`
	ID              string                   `json:"id"`
	Logo            *string                  `json:"logo"`
	Name            string                   `json:"name"`
	Pictures        []PlaylistChannelPicture `json:"pictures"`
	Slug            string                   `json:"slug"`
	Title           string                   `json:"title"`
	Type            string                   `json:"type"`
	TypeInternal    string                   `json:"__TYPE__"`
}

type PlaylistChannelGenericData struct {
	ID              string                   `json:"id"`
	Title           string                   `json:"title"`
	Name            string                   `json:"name"`
	Description     *string                  `json:"description"`
	Slug            string                   `json:"slug"`
	Target          string                   `json:"target"`
	Picture         string                   `json:"picture"`
	Pictures        []PlaylistChannelPicture `json:"pictures"`
	Type            string                   `json:"type"`
	TypeInternal    string                   `json:"__TYPE__"`
	BackgroundColor string                   `json:"background_color"`
	Logo            *string                  `json:"logo"`
}

type PlaylistChannelItemsType struct {
	ItemID           string                   `json:"item_id"`
	ID               string                   `json:"id"`
	Type             PlaylistChannelItemType  `json:"type"`
	Data             interface{}              `json:"data"`
	Target           string                   `json:"target"`
	Title            string                   `json:"title"`
	Subtitle         string                   `json:"subtitle"`
	Description      string                   `json:"description"`
	Pictures         []PlaylistChannelPicture `json:"pictures"`
	Weight           int                      `json:"weight"`
	LayoutParameters struct {
		CTA struct {
			Type  string `json:"type"`
			Label string `json:"label"`
		} `json:"cta"`
	} `json:"layout_parameters"`
}

func (p *PlaylistChannelItemsType) UnmarshalJSON(data []byte) error {
	type alias PlaylistChannelItemsType
	aux := struct {
		*alias
		Data json.RawMessage `json:"data"`
	}{
		alias: (*alias)(p),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if len(aux.Data) == 0 || string(aux.Data) == "null" {
		p.Data = nil
		return nil
	}

	switch p.Type {
	case PlaylistChannelItemPlaylist:
		var value PlaylistChannelPlaylistData
		if err := json.Unmarshal(aux.Data, &value); err != nil {
			return err
		}
		p.Data = &value
	case PlaylistChannelItemAlbum:
		var value PlaylistChannelAlbumData
		if err := json.Unmarshal(aux.Data, &value); err != nil {
			return err
		}
		p.Data = &value
	case PlaylistChannelItemArtist:
		var value ArtistInfoTypeMinimal
		if err := json.Unmarshal(aux.Data, &value); err != nil {
			return err
		}
		p.Data = &value
	case PlaylistChannelItemRadio:
		var value RadioType
		if err := json.Unmarshal(aux.Data, &value); err != nil {
			return err
		}
		p.Data = &value
	case PlaylistChannelItemShow:
		var value ShowType
		if err := json.Unmarshal(aux.Data, &value); err != nil {
			return err
		}
		p.Data = &value
	case PlaylistChannelItemEpisode:
		var value ShowEpisodeType
		if err := json.Unmarshal(aux.Data, &value); err != nil {
			return err
		}
		p.Data = &value
	case PlaylistChannelItemLivestream:
		var value PlaylistChannelLivestreamData
		if err := json.Unmarshal(aux.Data, &value); err != nil {
			return err
		}
		p.Data = &value
	case PlaylistChannelItemTrack:
		var value TrackType
		if err := json.Unmarshal(aux.Data, &value); err != nil {
			return err
		}
		p.Data = &value
	case PlaylistChannelItemNative:
		var value NativeAdsType
		if err := json.Unmarshal(aux.Data, &value); err != nil {
			return err
		}
		p.Data = &value
	case PlaylistChannelItemChannel, PlaylistChannelItemPage:
		var value PlaylistChannelGenericData
		if err := json.Unmarshal(aux.Data, &value); err != nil {
			return err
		}
		p.Data = &value
	default:
		var value map[string]interface{}
		if err := json.Unmarshal(aux.Data, &value); err != nil {
			return err
		}
		p.Data = value
	}

	return nil
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
