package metaflac

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"testing"
)

func TestNewMetaflac(t *testing.T) {
	// Generate a minimal FLAC buffer
	flacData := generateMinimalFlac()

	// Initialize Metaflac
	m, err := NewMetaflac(flacData)
	if err != nil {
		t.Fatalf("Failed to initialize Metaflac: %v", err)
	}

	if m == nil {
		t.Fatal("Metaflac instance is nil")
	}
}

func TestParseVorbisComment(t *testing.T) {
	// Prepare a sample Vorbis comment block
	vendorString := "test vendor"
	comments := []string{"TITLE=Test Song", "ARTIST=Test Artist"}
	vorbisCommentBlock := formatVorbisComment(vendorString, comments)

	// Initialize Metaflac with the Vorbis comment block
	m := &Metaflac{
		vorbisComment: vorbisCommentBlock,
	}
	err := m.parseVorbisComment()
	if err != nil {
		t.Fatalf("Failed to parse Vorbis comment: %v", err)
	}

	if m.GetVendorTag() != vendorString {
		t.Errorf("Expected vendor string '%s', got '%s'", vendorString, m.GetVendorTag())
	}

	if !equalStringSlices(m.GetAllTags(), comments) {
		t.Errorf("Expected tags %v, got %v", comments, m.GetAllTags())
	}
}

func TestGetMd5sum(t *testing.T) {
	// Prepare a sample STREAMINFO block with a known MD5 sum
	md5sumBytes, _ := hex.DecodeString("0123456789abcdef0123456789abcdef")
	streamInfo := make([]byte, 34)
	copy(streamInfo[18:], md5sumBytes)

	m := &Metaflac{
		streamInfo: streamInfo,
	}

	md5sum := m.GetMd5sum()
	if md5sum != "0123456789abcdef0123456789abcdef" {
		t.Errorf("Expected MD5 sum '0123456789abcdef0123456789abcdef', got '%s'", md5sum)
	}
}

func TestTagOperations(t *testing.T) {
	m := &Metaflac{
		tags: []string{"TITLE=Test Song", "ARTIST=Test Artist", "ALBUM=Test Album"},
	}

	// Test GetTag
	artistTags := m.GetTag("ARTIST")
	if len(artistTags) != 1 || artistTags[0] != "ARTIST=Test Artist" {
		t.Errorf("Expected 'ARTIST=Test Artist', got %v", artistTags)
	}

	// Test RemoveTag
	m.RemoveTag("ARTIST")
	if len(m.GetTag("ARTIST")) != 0 {
		t.Error("ARTIST tag was not removed")
	}

	// Test RemoveFirstTag
	m.SetTag("GENRE=Rock")
	m.SetTag("GENRE=Pop")
	m.RemoveFirstTag("GENRE")
	genreTags := m.GetTag("GENRE")
	if len(genreTags) != 1 || genreTags[0] != "GENRE=Pop" {
		t.Errorf("Expected 'GENRE=Pop', got %v", genreTags)
	}

	// Test RemoveAllTags
	m.RemoveAllTags()
	if len(m.GetAllTags()) != 0 {
		t.Error("All tags were not removed")
	}

	// Test SetTag
	err := m.SetTag("YEAR=2021")
	if err != nil {
		t.Errorf("Failed to set tag: %v", err)
	}
	if len(m.GetAllTags()) != 1 || m.GetAllTags()[0] != "YEAR=2021" {
		t.Errorf("Expected 'YEAR=2021', got %v", m.GetAllTags())
	}
}

func TestImportPicture(t *testing.T) {
	// Prepare sample picture data
	pictureData := []byte{0xFF, 0xD8, 0xFF} // Start of a JPEG file
	spec := PictureSpec{
		Type:        3,
		Mime:        "image/jpeg",
		Description: "Cover Art",
		Width:       600,
		Height:      600,
		Depth:       24,
		Colors:      0,
	}

	m := &Metaflac{}
	m.ImportPicture(pictureData, spec)

	if len(m.GetPicturesSpecs()) != 1 {
		t.Error("Picture was not imported correctly")
	}

	if !comparePictureSpec(m.GetPicturesSpecs()[0], spec) {
		t.Errorf("Expected picture spec %v, got %v", spec, m.GetPicturesSpecs()[0])
	}
}

func TestBuildPictureBlock(t *testing.T) {
	pictureData := []byte{0xFF, 0xD8, 0xFF}
	spec := PictureSpec{
		Type:        3,
		Mime:        "image/jpeg",
		Description: "Cover Art",
		Width:       600,
		Height:      600,
		Depth:       24,
		Colors:      0,
	}

	m := &Metaflac{}
	pictureBlock := m.buildPictureBlock(pictureData, spec)

	// Expected structure:
	// [Type][Mime Length][Mime][Description Length][Description]
	// [Width][Height][Depth][Colors][Picture Data Length][Picture Data]
	var buffer bytes.Buffer
	binary.Write(&buffer, binary.BigEndian, spec.Type)
	binary.Write(&buffer, binary.BigEndian, uint32(len(spec.Mime)))
	buffer.WriteString(spec.Mime)
	binary.Write(&buffer, binary.BigEndian, uint32(len(spec.Description)))
	buffer.WriteString(spec.Description)
	binary.Write(&buffer, binary.BigEndian, spec.Width)
	binary.Write(&buffer, binary.BigEndian, spec.Height)
	binary.Write(&buffer, binary.BigEndian, spec.Depth)
	binary.Write(&buffer, binary.BigEndian, spec.Colors)
	binary.Write(&buffer, binary.BigEndian, uint32(len(pictureData)))
	buffer.Write(pictureData)

	expected := buffer.Bytes()
	if !bytes.Equal(pictureBlock, expected) {
		t.Error("Picture block was not built correctly")
	}
}

func TestBuildMetadata(t *testing.T) {
	m := &Metaflac{
		streamInfo:   []byte{0x00, 0x01},
		blocks:       []Block{{BlockType: 3, Data: []byte{0x02, 0x03}}},
		tags:         []string{"TITLE=Test Song"},
		vendorString: "test vendor",
		pictures:     [][]byte{{0x04, 0x05}},
		picturesSpecs: []PictureSpec{{
			Type: 3, Mime: "image/jpeg",
		}},
	}

	metadata := m.buildMetadata()
	if len(metadata) != 5 {
		t.Errorf("Expected 5 metadata blocks, got %d", len(metadata))
	}

	// Verify STREAMINFO block
	streamInfoBlock := metadata[0]
	if streamInfoBlock[0]&0x7F != STREAMINFO {
		t.Error("First metadata block is not STREAMINFO")
	}

	// Verify VORBIS_COMMENT block
	vorbisCommentBlock := metadata[2]
	if vorbisCommentBlock[0]&0x7F != VORBIS_COMMENT {
		t.Error("Third metadata block is not VORBIS_COMMENT")
	}

	// Verify PICTURE block
	pictureBlock := metadata[3]
	if pictureBlock[0]&0x7F != PICTURE {
		t.Error("Fourth metadata block is not PICTURE")
	}

	// Verify PADDING block
	paddingBlock := metadata[4]
	if paddingBlock[0]&0x7F != PADDING {
		t.Error("Last metadata block is not PADDING")
	}
	if paddingBlock[0]&0x80 == 0 {
		t.Error("PADDING block is not marked as last")
	}
}

func TestBuildStream(t *testing.T) {
	originalData := generateMinimalFlac()
	m, err := NewMetaflac(originalData)
	if err != nil {
		t.Fatalf("Failed to initialize Metaflac: %v", err)
	}

	m.streamInfo = []byte{0x00, 0x01}
	m.tags = []string{"TITLE=Test Song"}
	m.vendorString = "test vendor"

	newStream := m.buildStream()
	if !bytes.HasPrefix(newStream, []byte("fLaC")) {
		t.Error("FLAC marker not present at the beginning of the stream")
	}
	if !bytes.Contains(newStream, []byte("Test Song")) {
		t.Error("Metadata not included in the stream")
	}
}

func TestGetBuffer(t *testing.T) {
	originalData := generateMinimalFlac()
	m, err := NewMetaflac(originalData)
	if err != nil {
		t.Fatalf("Failed to initialize Metaflac: %v", err)
	}

	m.streamInfo = []byte{0x00, 0x01}
	m.tags = []string{"TITLE=Test Song"}
	m.vendorString = "test vendor"

	newBuffer := m.GetBuffer()
	if !bytes.HasPrefix(newBuffer, []byte("fLaC")) {
		t.Error("FLAC marker not present at the beginning of the buffer")
	}
	if !bytes.Contains(newBuffer, []byte("Test Song")) {
		t.Error("Metadata not included in the buffer")
	}
}
