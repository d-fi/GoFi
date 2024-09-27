package metadata

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	id3v2 "github.com/bogem/id3v2/v2"
	"github.com/d-fi/GoFi/logger"
	"github.com/d-fi/GoFi/types"
)

func WriteMetadataMp3(buffer []byte, track types.TrackType, album *types.AlbumTypePublicApi, cover []byte) ([]byte, error) {
	logger.Debug("Starting MP3 metadata writing for track: %s", track.SNG_TITLE)

	// Create a reader from the buffer
	reader := bytes.NewReader(buffer)

	// Parse the tag from the reader
	tag, err := id3v2.ParseReader(reader, id3v2.Options{Parse: true})
	var audioData []byte
	if err != nil {
		// No existing tag or error reading, create a new one
		tag = id3v2.NewEmptyTag()
		audioData = buffer
		logger.Debug("No existing tags found, creating new tag")
	} else {
		// Read the rest of the reader as audio data
		audioData, err = io.ReadAll(reader)
		if err != nil {
			logger.Debug("Failed to read audio data: %v", err)
			return nil, err
		}
		logger.Debug("Existing tags found, extracted audio data")
	}

	// Set tag version to ID3v2.3 for better compatibility
	tag.SetVersion(3)

	// Set default encoding to UTF-16 for ID3v2.3 compatibility
	tag.SetDefaultEncoding(id3v2.EncodingUTF16)

	// Set frames
	tag.SetTitle(track.SNG_TITLE)
	tag.SetAlbum(track.ALB_TITLE)

	// Set artist names
	var artistNames []string
	for _, artist := range track.ARTISTS {
		artistNames = append(artistNames, artist.ART_NAME)
	}
	tag.SetArtist(strings.Join(artistNames, ", "))
	logger.Debug("Set basic MP3 tags: TITLE, ALBUM, ARTIST")

	// TLEN (Length)
	durationMs := int(track.DURATION) * 1000
	durationFrame := id3v2.TextFrame{
		Encoding: id3v2.EncodingUTF16,
		Text:     fmt.Sprintf("%d", durationMs),
	}
	tag.AddFrame(tag.CommonID("Length"), durationFrame)

	// TSRC (ISRC)
	isrcFrame := id3v2.TextFrame{
		Encoding: id3v2.EncodingUTF16,
		Text:     track.ISRC,
	}
	tag.AddFrame(tag.CommonID("ISRC"), isrcFrame)
	logger.Debug("Set duration and ISRC")

	// Album-related frames
	if album != nil {
		if len(album.Genres.Data) > 0 {
			var genres []string
			for _, genre := range album.Genres.Data {
				genres = append(genres, genre.Name)
			}
			tag.SetGenre(strings.Join(genres, ", "))
		}

		releaseDates := strings.Split(album.ReleaseDate, "-")
		if len(releaseDates) >= 1 {
			tag.SetYear(releaseDates[0])
		}
		if len(releaseDates) >= 3 {
			date := releaseDates[2] + releaseDates[1]
			dateFrame := id3v2.TextFrame{
				Encoding: id3v2.EncodingUTF16,
				Text:     date,
			}
			tag.AddFrame("TDAT", dateFrame)
		}

		albumArtistFrame := id3v2.TextFrame{
			Encoding: id3v2.EncodingUTF16,
			Text:     album.Artist.Name,
		}
		tag.AddFrame("TPE2", albumArtistFrame)

		addUserTextFrame(tag, "RELEASETYPE", album.RecordType)
		addUserTextFrame(tag, "BARCODE", album.UPC)
		addUserTextFrame(tag, "LABEL", album.Label)

		compilation := "0"
		if strings.Contains(strings.ToLower(album.Artist.Name), "various") {
			compilation = "1"
		}
		addUserTextFrame(tag, "COMPILATION", compilation)
		logger.Debug("Set album-related tags and frames")
	}

	mediaTypeFrame := id3v2.TextFrame{
		Encoding: id3v2.EncodingUTF16,
		Text:     "Digital Media",
	}
	tag.AddFrame("TMED", mediaTypeFrame)
	addUserTextFrame(tag, "SOURCE", "Deezer")
	addUserTextFrame(tag, "SOURCEID", track.SNG_ID)

	if track.DISK_NUMBER != 0 {
		partOfSetFrame := id3v2.TextFrame{
			Encoding: id3v2.EncodingUTF16,
			Text:     fmt.Sprintf("%d", int(track.DISK_NUMBER)),
		}
		tag.AddFrame("TPOS", partOfSetFrame)
	}

	trackNumber := fmt.Sprintf("%02d", int(track.TRACK_NUMBER))
	if album != nil {
		totalTracks := fmt.Sprintf("%02d", album.NbTracks)
		trackFrame := id3v2.TextFrame{
			Encoding: id3v2.EncodingUTF16,
			Text:     fmt.Sprintf("%s/%s", trackNumber, totalTracks),
		}
		tag.AddFrame("TRCK", trackFrame)
	} else {
		trackFrame := id3v2.TextFrame{
			Encoding: id3v2.EncodingUTF16,
			Text:     trackNumber,
		}
		tag.AddFrame("TRCK", trackFrame)
	}

	// Set contributors
	if track.SNG_CONTRIBUTORS != nil {
		contributors := track.SNG_CONTRIBUTORS
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
				Encoding: id3v2.EncodingUTF16,
				Text:     copyright,
			}
			tag.AddFrame("TCOP", copyrightFrame)
		}
		if len(contributors.Publisher) > 0 {
			publisherFrame := id3v2.TextFrame{
				Encoding: id3v2.EncodingUTF16,
				Text:     strings.Join(contributors.Publisher, ", "),
			}
			tag.AddFrame("TPUB", publisherFrame)
		}
		if len(contributors.Composer) > 0 {
			composerFrame := id3v2.TextFrame{
				Encoding: id3v2.EncodingUTF16,
				Text:     strings.Join(contributors.Composer, ", "),
			}
			tag.AddFrame("TCOM", composerFrame)
		}

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
		logger.Debug("Set contributor-related tags")
	}

	// Set lyrics and explicit content
	if track.LYRICS != nil {
		lyricsFrame := id3v2.UnsynchronisedLyricsFrame{
			Encoding:          id3v2.EncodingUTF16,
			Language:          "eng",
			ContentDescriptor: "",
			Lyrics:            track.LYRICS.LYRICS_TEXT,
		}
		tag.AddUnsynchronisedLyricsFrame(lyricsFrame)
	}
	if track.EXPLICIT_LYRICS != nil {
		addUserTextFrame(tag, "EXPLICIT", fmt.Sprintf("%t", *track.EXPLICIT_LYRICS))
	}

	// Add cover art
	if cover != nil {
		picture := id3v2.PictureFrame{
			Encoding:    id3v2.EncodingISO, // Use ISO encoding for PictureFrame in ID3v2.3
			MimeType:    "image/jpeg",
			PictureType: id3v2.PTFrontCover,
			Description: "",
			Picture:     cover,
		}
		tag.AddAttachedPicture(picture)
		logger.Debug("Added cover art to MP3 metadata")
	}

	// Write the tag and audio data to a new buffer
	var newBuffer bytes.Buffer
	if _, err := tag.WriteTo(&newBuffer); err != nil {
		logger.Debug("Failed to write MP3 tags: %v", err)
		return nil, err
	}

	// Append the audio data
	newBuffer.Write(audioData)
	logger.Debug("Completed MP3 metadata writing for track: %s", track.SNG_TITLE)

	return newBuffer.Bytes(), nil
}

// Helper function to add a TXXX frame
func addUserTextFrame(tag *id3v2.Tag, description, value string) {
	frame := id3v2.UserDefinedTextFrame{
		Encoding:    id3v2.EncodingUTF16,
		Description: description,
		Value:       value,
	}
	tag.AddFrame(tag.CommonID("UserDefinedText"), frame)
	logger.Debug("Added TXXX frame: %s = %s", description, value)
}
