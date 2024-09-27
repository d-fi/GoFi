package metadata

import (
	"fmt"
	"strings"

	"github.com/d-fi/GoFi/metaflac"
	"github.com/d-fi/GoFi/types"
)

func WriteMetadataFlac(buffer []byte, track types.TrackType, album *types.AlbumTypePublicApi, dimension int, cover []byte) ([]byte, error) {
	flac, err := metaflac.NewMetaflac(buffer)
	if err != nil {
		return nil, err
	}

	var RELEASE_YEAR string
	if album != nil {
		RELEASE_YEAR = strings.Split(album.ReleaseDate, "-")[0]
	}

	flac.SetTag("TITLE=" + track.SNG_TITLE)
	flac.SetTag("ALBUM=" + track.ALB_TITLE)

	var artistNames []string
	for _, artist := range track.ARTISTS {
		artistNames = append(artistNames, artist.ART_NAME)
	}
	flac.SetTag("ARTIST=" + strings.Join(artistNames, ", "))

	flac.SetTag(fmt.Sprintf("TRACKNUMBER=%02d", int(track.TRACK_NUMBER)))

	if album != nil {
		TOTALTRACKS := fmt.Sprintf("%02d", album.NbTracks)

		if len(album.Genres.Data) > 0 {
			for _, genre := range album.Genres.Data {
				flac.SetTag("GENRE=" + genre.Name)
			}
		}

		flac.SetTag("TRACKTOTAL=" + TOTALTRACKS)
		flac.SetTag("TOTALTRACKS=" + TOTALTRACKS)
		flac.SetTag("RELEASETYPE=" + album.RecordType)
		flac.SetTag("ALBUMARTIST=" + album.Artist.Name)
		flac.SetTag("BARCODE=" + album.UPC)
		flac.SetTag("LABEL=" + album.Label)
		flac.SetTag("DATE=" + album.ReleaseDate)
		flac.SetTag("YEAR=" + RELEASE_YEAR)

		compilation := "0"
		if strings.Contains(strings.ToLower(album.Artist.Name), "various") {
			compilation = "1"
		}
		flac.SetTag("COMPILATION=" + compilation)
	}

	if track.DISK_NUMBER != 0 {
		flac.SetTag(fmt.Sprintf("DISCNUMBER=%d", int(track.DISK_NUMBER)))
	}

	flac.SetTag("ISRC=" + track.ISRC)
	flac.SetTag(fmt.Sprintf("LENGTH=%d", int(track.DURATION)))
	flac.SetTag("MEDIA=Digital Media")

	if track.LYRICS != nil {
		flac.SetTag("LYRICS=" + track.LYRICS.LYRICS_TEXT)
	}

	if track.EXPLICIT_LYRICS != nil {
		flac.SetTag(fmt.Sprintf("EXPLICIT=%t", bool(*track.EXPLICIT_LYRICS)))
	}

	if track.SNG_CONTRIBUTORS != nil {
		contributors := track.SNG_CONTRIBUTORS

		if len(contributors.MainArtist) > 0 {
			copyright := RELEASE_YEAR
			if RELEASE_YEAR != "" {
				copyright += " "
			}
			copyright += contributors.MainArtist[0]
			flac.SetTag("COPYRIGHT=" + copyright)
		}
		if len(contributors.Publisher) > 0 {
			flac.SetTag("ORGANIZATION=" + strings.Join(contributors.Publisher, ", "))
		}
		if len(contributors.Composer) > 0 {
			flac.SetTag("COMPOSER=" + strings.Join(contributors.Composer, ", "))
		}
		if len(contributors.Producer) > 0 {
			flac.SetTag("PRODUCER=" + strings.Join(contributors.Producer, ", "))
		}
		if len(contributors.Engineer) > 0 {
			flac.SetTag("ENGINEER=" + strings.Join(contributors.Engineer, ", "))
		}
		if len(contributors.Writer) > 0 {
			flac.SetTag("WRITER=" + strings.Join(contributors.Writer, ", "))
		}
		if len(contributors.Author) > 0 {
			flac.SetTag("AUTHOR=" + strings.Join(contributors.Author, ", "))
		}
		if len(contributors.Mixer) > 0 {
			flac.SetTag("MIXER=" + strings.Join(contributors.Mixer, ", "))
		}
	}

	if cover != nil {
		spec := metaflac.PictureSpec{
			Type:        3,
			Mime:        "image/jpeg",
			Description: "",
			Width:       uint32(dimension),
			Height:      uint32(dimension),
			Depth:       24,
			Colors:      0,
		}
		flac.ImportPicture(cover, spec)
	}

	flac.SetTag("SOURCE=Deezer")
	flac.SetTag("SOURCEID=" + track.SNG_ID)

	newBuffer := flac.GetBuffer()
	return newBuffer, nil
}