package metadata

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/bogem/id3v2/v2"
	"github.com/d-fi/GoFi/types"
)

func WriteMetadataMp3(buffer []byte, track types.TrackType, album *types.AlbumTypePublicApi, cover []byte) ([]byte, error) {
	// Create a reader from the buffer
	reader := bytes.NewReader(buffer)

	// Parse the tag from the reader
	tag, err := id3v2.ParseReader(reader, id3v2.Options{Parse: true})
	var audioData []byte
	if err != nil {
		// No existing tag or error reading, create a new one
		tag = id3v2.NewEmptyTag()
		audioData = buffer
	} else {
		// Existing tag found, extract the audio data
		tagSize := int(tag.Size())
		if tagSize > len(buffer) {
			tagSize = len(buffer)
		}
		audioData = buffer[tagSize:]
	}

	// Set frames

	// TIT2 (Title/songname/content description)
	tag.SetTitle(track.SNG_TITLE)

	// TALB (Album/Movie/Show title)
	tag.SetAlbum(track.ALB_TITLE)

	// TPE1 (Lead performer(s)/Soloist(s))
	var artistNames []string
	for _, artist := range track.ARTISTS {
		artistNames = append(artistNames, artist.ART_NAME)
	}
	tag.SetArtist(strings.Join(artistNames, ", "))

	// TLEN (Length)
	durationMs := int(track.DURATION) * 1000
	durationFrame := id3v2.TextFrame{
		Encoding: id3v2.EncodingUTF8,
		Text:     fmt.Sprintf("%d", durationMs),
	}
	tag.AddFrame(tag.CommonID("Length"), durationFrame)

	// TSRC (ISRC)
	isrcFrame := id3v2.TextFrame{
		Encoding: id3v2.EncodingUTF8,
		Text:     track.ISRC,
	}
	tag.AddFrame(tag.CommonID("ISRC"), isrcFrame)

	// Album-related frames
	if album != nil {
		// Genres (TCON)
		if len(album.Genres.Data) > 0 {
			var genres []string
			for _, genre := range album.Genres.Data {
				genres = append(genres, genre.Name)
			}
			tag.SetGenre(strings.Join(genres, ", "))
		}

		// Release date components
		releaseDates := strings.Split(album.ReleaseDate, "-")
		if len(releaseDates) >= 1 {
			// TYER (Year)
			tag.SetYear(releaseDates[0])
		}
		if len(releaseDates) >= 3 {
			// TDAT (Date DDMM)
			date := releaseDates[2] + releaseDates[1] // DDMM
			dateFrame := id3v2.TextFrame{
				Encoding: id3v2.EncodingUTF8,
				Text:     date,
			}
			tag.AddFrame("TDAT", dateFrame)
		}

		// TPE2 (Album artist)
		albumArtistFrame := id3v2.TextFrame{
			Encoding: id3v2.EncodingUTF8,
			Text:     album.Artist.Name,
		}
		tag.AddFrame("TPE2", albumArtistFrame)

		// Custom TXXX frames
		addUserTextFrame(tag, "RELEASETYPE", album.RecordType)
		addUserTextFrame(tag, "BARCODE", album.UPC)
		addUserTextFrame(tag, "LABEL", album.Label)

		compilation := "0"
		if strings.Contains(strings.ToLower(album.Artist.Name), "various") {
			compilation = "1"
		}
		addUserTextFrame(tag, "COMPILATION", compilation)
	}

	// TMED (Media type)
	mediaTypeFrame := id3v2.TextFrame{
		Encoding: id3v2.EncodingUTF8,
		Text:     "Digital Media",
	}
	tag.AddFrame("TMED", mediaTypeFrame)

	// Source TXXX frames
	addUserTextFrame(tag, "SOURCE", "Deezer")
	addUserTextFrame(tag, "SOURCEID", track.SNG_ID)

	// Disk number and track number
	if track.DISK_NUMBER != 0 {
		// TPOS (Part of a set)
		partOfSetFrame := id3v2.TextFrame{
			Encoding: id3v2.EncodingUTF8,
			Text:     fmt.Sprintf("%d", int(track.DISK_NUMBER)),
		}
		tag.AddFrame("TPOS", partOfSetFrame)
	}

	// TRCK (Track number)
	trackNumber := fmt.Sprintf("%02d", int(track.TRACK_NUMBER))
	if album != nil {
		totalTracks := fmt.Sprintf("%02d", album.NbTracks)
		trackFrame := id3v2.TextFrame{
			Encoding: id3v2.EncodingUTF8,
			Text:     fmt.Sprintf("%s/%s", trackNumber, totalTracks),
		}
		tag.AddFrame("TRCK", trackFrame)
	} else {
		trackFrame := id3v2.TextFrame{
			Encoding: id3v2.EncodingUTF8,
			Text:     trackNumber,
		}
		tag.AddFrame("TRCK", trackFrame)
	}

	// Contributors
	if track.SNG_CONTRIBUTORS != nil {
		contributors := track.SNG_CONTRIBUTORS

		// TCOP (Copyright)
		if len(contributors.MainArtist) > 0 {
			var releaseYear string
			if album != nil {
				releaseDates := strings.Split(album.ReleaseDate, "-")
				if len(releaseDates) >= 1 {
					releaseYear = releaseDates[0]
				}
			}
			copyright := fmt.Sprintf("%s %s", releaseYear, contributors.MainArtist[0])
			copyrightFrame := id3v2.TextFrame{
				Encoding: id3v2.EncodingUTF8,
				Text:     copyright,
			}
			tag.AddFrame("TCOP", copyrightFrame)
		}

		// TPUB (Publisher)
		if len(contributors.Publisher) > 0 {
			publisherFrame := id3v2.TextFrame{
				Encoding: id3v2.EncodingUTF8,
				Text:     strings.Join(contributors.Publisher, ", "),
			}
			tag.AddFrame("TPUB", publisherFrame)
		}

		// TCOM (Composer)
		if len(contributors.Composer) > 0 {
			composerFrame := id3v2.TextFrame{
				Encoding: id3v2.EncodingUTF8,
				Text:     strings.Join(contributors.Composer, ", "),
			}
			tag.AddFrame("TCOM", composerFrame)
		}

		// Additional TXXX frames for contributors
		if len(contributors.Writer) > 0 {
			addUserTextFrame(tag, "LYRICIST", strings.Join(contributors.Writer, ", "))
		}
		if len(contributors.Author) > 0 {
			addUserTextFrame(tag, "AUTHOR", strings.Join(contributors.Author, ", "))
		}
		if len(contributors.Mixer) > 0 {
			addUserTextFrame(tag, "MIXARTIST", strings.Join(contributors.Mixer, ", "))
		}
		if len(contributors.Producer) > 0 && len(contributors.Engineer) > 0 {
			involvedPeople := append(contributors.Producer, contributors.Engineer...)
			addUserTextFrame(tag, "INVOLVEDPEOPLE", strings.Join(involvedPeople, ", "))
		}
	}

	// Lyrics (USLT)
	if track.LYRICS != nil {
		lyricsFrame := id3v2.UnsynchronisedLyricsFrame{
			Encoding:          id3v2.EncodingUTF8,
			Language:          "eng",
			ContentDescriptor: "",
			Lyrics:            track.LYRICS.LYRICS_TEXT,
		}
		tag.AddUnsynchronisedLyricsFrame(lyricsFrame)
	}

	// Explicit lyrics TXXX frame
	if track.EXPLICIT_LYRICS != nil {
		addUserTextFrame(tag, "EXPLICIT", fmt.Sprintf("%t", *track.EXPLICIT_LYRICS))
	}

	// Cover art (APIC)
	if cover != nil {
		picture := id3v2.PictureFrame{
			Encoding:    id3v2.EncodingUTF8,
			MimeType:    "image/jpeg",
			PictureType: id3v2.PTFrontCover,
			Description: "",
			Picture:     cover,
		}
		tag.AddAttachedPicture(picture)
	}

	// Write the tag and audio data to a new buffer
	var newBuffer bytes.Buffer
	if _, err := tag.WriteTo(&newBuffer); err != nil {
		return nil, err
	}
	// Append the audio data
	newBuffer.Write(audioData)

	return newBuffer.Bytes(), nil
}

// Helper function to add a TXXX frame
func addUserTextFrame(tag *id3v2.Tag, description, value string) {
	frame := id3v2.UserDefinedTextFrame{
		Encoding:    id3v2.EncodingUTF8,
		Description: description,
		Value:       value,
	}
	tag.AddFrame(tag.CommonID("UserDefinedText"), frame)
}
