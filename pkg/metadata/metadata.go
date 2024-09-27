package metadata

import (
	"bytes"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/d-fi/GoFi/pkg/api"
	"github.com/d-fi/GoFi/pkg/types"
)

// AddTrackTags adds metadata to the track buffer (MP3 or FLAC) based on track and album information.
func AddTrackTags(trackBuffer []byte, track types.TrackType, albumCoverSize int) ([]byte, error) {
	cover, coverErr := DownloadAlbumCover(track.ALB_PICTURE, albumCoverSize)
	if coverErr != nil {
		return nil, coverErr
	}

	var lyrics types.LyricsType
	if track.LYRICS_ID > 0 {
		var lyricsErr error
		lyrics, lyricsErr = api.GetLyrics(track.SNG_ID)
		if lyricsErr == nil {
			track.LYRICS = &lyrics
		}
	}

	album, albumErr := api.GetAlbumInfoPublicApi(track.ALB_ID)
	if albumErr != nil {
		return nil, albumErr
	}

	if strings.ToLower(track.ART_NAME) == "various" {
		track.ART_NAME = "Various Artists"
	}

	if album.RecordType != "" {
		caser := cases.Title(language.English)
		if strings.ToLower(album.RecordType) == "ep" {
			album.RecordType = "EP"
		} else {
			album.RecordType = caser.String(album.RecordType)
		}
	}

	isFlac := bytes.HasPrefix(trackBuffer, []byte("fLaC"))
	if isFlac {
		return WriteMetadataFlac(trackBuffer, track, &album, albumCoverSize, cover)
	}

	return WriteMetadataMp3(trackBuffer, track, &album, cover)
}
