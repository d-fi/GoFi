package metadata

import (
	"bytes"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/d-fi/GoFi/api"
	"github.com/d-fi/GoFi/logger"
	"github.com/d-fi/GoFi/types"
)

// AddTrackTags adds metadata to the track buffer (MP3 or FLAC) based on track and album information.
func AddTrackTags(trackBuffer []byte, track types.TrackType, albumCoverSize int) ([]byte, error) {
	logger.Debug("Starting to add track tags for track: %s", track.SNG_TITLE)

	cover, coverErr := DownloadAlbumCover(track.ALB_PICTURE, albumCoverSize)
	if coverErr != nil {
		logger.Debug("Failed to download album cover: %v", coverErr)
		return nil, coverErr
	}
	logger.Debug("Downloaded album cover successfully")

	var lyrics types.LyricsType
	if track.LYRICS_ID > 0 {
		var lyricsErr error
		lyrics, lyricsErr = api.GetLyrics(track.SNG_ID)
		if lyricsErr == nil {
			track.LYRICS = &lyrics
			logger.Debug("Fetched lyrics successfully for track: %s", track.SNG_TITLE)
		} else {
			logger.Debug("Failed to fetch lyrics: %v", lyricsErr)
		}
	}

	album, albumErr := api.GetAlbumInfoPublicApi(track.ALB_ID)
	if albumErr != nil {
		logger.Debug("Failed to fetch album info: %v", albumErr)
		return nil, albumErr
	}
	logger.Debug("Fetched album info successfully for album: %s", album.Title)

	if strings.ToLower(track.ART_NAME) == "various" {
		track.ART_NAME = "Various Artists"
		logger.Debug("Adjusted artist name to 'Various Artists'")
	}

	if album.RecordType != "" {
		caser := cases.Title(language.English)
		if strings.ToLower(album.RecordType) == "ep" {
			album.RecordType = "EP"
		} else {
			album.RecordType = caser.String(album.RecordType)
		}
		logger.Debug("Formatted album record type: %s", album.RecordType)
	}

	isFlac := bytes.HasPrefix(trackBuffer, []byte("fLaC"))
	if isFlac {
		logger.Debug("Detected FLAC format for track: %s", track.SNG_TITLE)
		return WriteMetadataFlac(trackBuffer, track, &album, albumCoverSize, cover)
	}

	logger.Debug("Detected MP3 format for track: %s", track.SNG_TITLE)
	return WriteMetadataMp3(trackBuffer, track, &album, cover)
}
