package metadata

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/bogem/id3v2/v2"
	"github.com/d-fi/GoFi/logger"
	"github.com/d-fi/GoFi/types"
)

// WriteMetadataMp3 writes metadata to an MP3 buffer and returns the updated buffer.
func WriteMetadataMp3(buffer []byte, track types.TrackType, album *types.AlbumTypePublicApi, cover []byte) ([]byte, error) {
	logger.Debug("Starting MP3 metadata writing for track: %s", track.SNG_TITLE)

	reader := bytes.NewReader(buffer)
	tag, err := id3v2.ParseReader(reader, id3v2.Options{Parse: true})
	var audioData []byte
	if err != nil {
		// If no existing tags, create a new tag
		tag = id3v2.NewEmptyTag()
		audioData = buffer
		logger.Debug("No existing tags found, creating new tag")
	} else {
		// Extract audio data after tags
		audioData, err = io.ReadAll(reader)
		if err != nil {
			logger.Debug("Failed to read audio data: %v", err)
			return nil, err
		}
		logger.Debug("Existing tags found, extracted audio data")
	}

	// Set ID3 version to 2.4
	tag.SetVersion(4)
	// Set default encoding to UTF-16
	tag.SetDefaultEncoding(id3v2.EncodingUTF16)

	// Set standard frames
	tag.SetTitle(track.SNG_TITLE)
	tag.SetAlbum(track.ALB_TITLE)
	tag.SetArtist(strings.Join(processArtistNames(track.ARTISTS), "/"))
	tag.AddTextFrame("TLEN", id3v2.EncodingUTF16, fmt.Sprintf("%d", track.DURATION*1000)) // TLEN expects milliseconds
	tag.AddTextFrame("TSRC", id3v2.EncodingUTF16, track.ISRC)

	// Set album metadata if available
	if album != nil {
		setAlbumMetadata(tag, album)
	}

	// Set additional frames
	tag.AddTextFrame("TMED", id3v2.EncodingUTF16, "Digital Media")
	addUserTextFrame(tag, "SOURCE", "Deezer")
	addUserTextFrame(tag, "SOURCEID", track.SNG_ID)

	// Set track number and disc number if available
	if track.DISK_NUMBER != 0 {
		setTrackNumberFrames(tag, track, album)
	}

	// Set contributors metadata
	setContributorsMetadata(tag, track, album)

	// Set lyrics and explicit lyrics
	if track.LYRICS != nil {
		tag.AddUnsynchronisedLyricsFrame(id3v2.UnsynchronisedLyricsFrame{
			Encoding: id3v2.EncodingUTF16,
			Language: "eng",
			Lyrics:   track.LYRICS.LYRICS_TEXT,
		})
	}
	if track.EXPLICIT_LYRICS != nil {
		addUserTextFrame(tag, "EXPLICIT", fmt.Sprintf("%t", *track.EXPLICIT_LYRICS))
	}

	// Add cover art if available
	if cover != nil {
		tag.AddAttachedPicture(id3v2.PictureFrame{
			Encoding:    id3v2.EncodingUTF16,
			MimeType:    "image/jpeg",
			PictureType: id3v2.PTFrontCover,
			Description: "",
			Picture:     cover,
		})
		logger.Debug("Added cover art to MP3 metadata")
	}

	// Write the tag to a new buffer
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

// processArtistNames splits artist names by slash, trims them, and joins with slash without spaces.
func processArtistNames(artists []types.ArtistType) []string {
	var names []string
	for _, artist := range artists {
		splitNames := strings.Split(artist.ART_NAME, "/")
		for _, name := range splitNames {
			trimmedName := strings.TrimSpace(name)
			if trimmedName != "" {
				names = append(names, trimmedName)
			}
		}
	}
	return names
}

// setAlbumMetadata sets album-related metadata frames.
func setAlbumMetadata(tag *id3v2.Tag, album *types.AlbumTypePublicApi) {
	// Set genre(s) if available
	if len(album.Genres.Data) > 0 {
		var genres []string
		for _, genre := range album.Genres.Data {
			genres = append(genres, genre.Name)
		}
		tag.SetGenre(strings.Join(genres, ", "))
	}

	// Set recording year using TDRC and TYER
	releaseDates := strings.Split(album.ReleaseDate, "-")
	if len(releaseDates) >= 1 {
		year := releaseDates[0]
		tag.AddTextFrame("TDRC", id3v2.EncodingUTF16, year) // Recording year
		tag.AddTextFrame("TYER", id3v2.EncodingUTF16, year) // Recording year for compatibility
	}
	if len(releaseDates) >= 3 {
		tag.AddTextFrame("TDAT", id3v2.EncodingUTF16, releaseDates[2]+releaseDates[1])
	}

	// Set album artist
	tag.AddTextFrame("TPE2", id3v2.EncodingUTF16, album.Artist.Name)

	// Set custom frames
	addUserTextFrame(tag, "RELEASETYPE", album.RecordType)
	addUserTextFrame(tag, "BARCODE", album.UPC)
	addUserTextFrame(tag, "LABEL", album.Label)
	addUserTextFrame(tag, "COMPILATION", ifMatchVarious(album.Artist.Name))
}

// setTrackNumberFrames sets TRCK and TPOS frames.
func setTrackNumberFrames(tag *id3v2.Tag, track types.TrackType, album *types.AlbumTypePublicApi) {
	trackNumber := fmt.Sprintf("%02d", int(track.TRACK_NUMBER))
	if album != nil {
		totalTracks := fmt.Sprintf("%02d", album.NbTracks)
		tag.AddTextFrame("TRCK", id3v2.EncodingUTF16, fmt.Sprintf("%s/%s", trackNumber, totalTracks))
	} else {
		tag.AddTextFrame("TRCK", id3v2.EncodingUTF16, trackNumber)
	}
	tag.AddTextFrame("TPOS", id3v2.EncodingUTF16, fmt.Sprintf("%d", int(track.DISK_NUMBER)))
}

// setContributorsMetadata sets contributor-related metadata frames.
func setContributorsMetadata(tag *id3v2.Tag, track types.TrackType, album *types.AlbumTypePublicApi) {
	contributors := track.SNG_CONTRIBUTORS
	if contributors == nil {
		return
	}

	// Set TCOP (Copyright)
	if len(contributors.MainArtist) > 0 {
		releaseYear := ""
		if album != nil {
			releaseDates := strings.Split(album.ReleaseDate, "-")
			if len(releaseDates) >= 1 {
				releaseYear = releaseDates[0]
			}
		}
		tag.AddTextFrame("TCOP", id3v2.EncodingUTF16, fmt.Sprintf("%s %s", releaseYear, contributors.MainArtist[0]))
	}

	// Set other contributor frames
	addContributorsFrames(tag, contributors)
}

// addContributorsFrames adds various contributor frames.
func addContributorsFrames(tag *id3v2.Tag, contributors *types.SongContributors) {
	if len(contributors.Publisher) > 0 {
		tag.AddTextFrame("TPUB", id3v2.EncodingUTF16, strings.Join(contributors.Publisher, ", "))
	}
	if len(contributors.Composer) > 0 {
		tag.AddTextFrame("TCOM", id3v2.EncodingUTF16, strings.Join(contributors.Composer, "/"))
	}
	if len(contributors.Writer) > 0 {
		addUserTextFrame(tag, "LYRICIST", strings.Join(contributors.Writer, "/"))
	}
	if len(contributors.Author) > 0 {
		addUserTextFrame(tag, "AUTHOR", strings.Join(contributors.Author, "/"))
	}
	if len(contributors.Mixer) > 0 {
		addUserTextFrame(tag, "MIXARTIST", strings.Join(contributors.Mixer, "/"))
	}
	if len(contributors.Producer) > 0 && len(contributors.Engineer) > 0 {
		involvedPeople := append(contributors.Producer, contributors.Engineer...)
		addUserTextFrame(tag, "INVOLVEDPEOPLE", strings.Join(involvedPeople, "/"))
	}
}

// ifMatchVarious returns "1" if artist name contains "various", else "0".
func ifMatchVarious(artistName string) string {
	if strings.Contains(strings.ToLower(artistName), "various") {
		return "1"
	}
	return "0"
}

// addUserTextFrame adds a TXXX frame with the given description and value.
func addUserTextFrame(tag *id3v2.Tag, description, value string) {
	tag.AddUserDefinedTextFrame(id3v2.UserDefinedTextFrame{
		Encoding:    id3v2.EncodingUTF16,
		Description: description,
		Value:       value,
	})
	logger.Debug("Added TXXX frame: %s = %s", description, value)
}
