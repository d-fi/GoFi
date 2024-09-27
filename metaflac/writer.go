package metaflac

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"strings"
)

// Metadata block types
const (
	STREAMINFO     = 0
	PADDING        = 1
	APPLICATION    = 2
	SEEKTABLE      = 3
	VORBIS_COMMENT = 4
	CUESHEET       = 5
	PICTURE        = 6
)

// Block represents a FLAC metadata block.
type Block struct {
	BlockType int
	Data      []byte
}

// PictureSpec holds the specifications of a picture block.
type PictureSpec struct {
	Type        uint32
	Mime        string
	Description string
	Width       uint32
	Height      uint32
	Depth       uint32
	Colors      uint32
}

// Metaflac handles FLAC metadata manipulation.
type Metaflac struct {
	buffer        []byte
	marker        string
	streamInfo    []byte
	blocks        []Block
	padding       []byte
	vorbisComment []byte
	vendorString  string
	tags          []string
	pictures      [][]byte
	picturesSpecs []PictureSpec
	picturesDatas [][]byte
	framesOffset  int
}

// NewMetaflac initializes a new Metaflac instance.
func NewMetaflac(flac []byte) (*Metaflac, error) {
	m := &Metaflac{
		buffer:        flac,
		blocks:        []Block{},
		tags:          []string{},
		pictures:      [][]byte{},
		picturesSpecs: []PictureSpec{},
		picturesDatas: [][]byte{},
	}
	if err := m.init(); err != nil {
		return nil, err
	}
	return m, nil
}

// Initialize by parsing the FLAC file's metadata blocks.
func (m *Metaflac) init() error {
	if len(m.buffer) < 4 || string(m.buffer[:4]) != "fLaC" {
		return errors.New("invalid FLAC file")
	}
	offset := 4
	isLastBlock := false

	for !isLastBlock {
		if offset >= len(m.buffer) {
			return errors.New("unexpected end of file")
		}

		blockHeader := m.buffer[offset]
		offset++

		isLastBlock = blockHeader&0x80 != 0
		blockType := int(blockHeader & 0x7F)

		if offset+3 > len(m.buffer) {
			return errors.New("unexpected end of file")
		}
		blockLength := int(binary.BigEndian.Uint32(append([]byte{0}, m.buffer[offset:offset+3]...)))
		offset += 3

		if offset+blockLength > len(m.buffer) {
			return errors.New("unexpected end of file")
		}

		blockData := m.buffer[offset : offset+blockLength]

		switch blockType {
		case STREAMINFO:
			m.streamInfo = blockData
		case VORBIS_COMMENT:
			m.vorbisComment = blockData
			m.parseVorbisComment()
		case APPLICATION, SEEKTABLE, CUESHEET:
			m.blocks = append(m.blocks, Block{BlockType: blockType, Data: blockData})
		case PICTURE:
			m.pictures = append(m.pictures, blockData)
			m.parsePictureBlock(blockData)
		}

		offset += blockLength
	}
	m.framesOffset = offset
	return nil
}

// Parse the Vorbis comment block to extract the vendor string and tags.
func (m *Metaflac) parseVorbisComment() error {
	if len(m.vorbisComment) < 4 {
		return errors.New("invalid Vorbis comment block")
	}
	vendorLength := binary.LittleEndian.Uint32(m.vorbisComment[:4])
	if int(4+vendorLength) > len(m.vorbisComment) {
		return errors.New("invalid vendor length in Vorbis comment")
	}
	m.vendorString = string(m.vorbisComment[4 : 4+vendorLength])

	offset := int(4 + vendorLength)
	if offset+4 > len(m.vorbisComment) {
		return errors.New("invalid Vorbis comment block")
	}
	userCommentListLength := binary.LittleEndian.Uint32(m.vorbisComment[offset : offset+4])
	offset += 4

	for i := uint32(0); i < userCommentListLength; i++ {
		if offset+4 > len(m.vorbisComment) {
			return errors.New("invalid Vorbis comment block")
		}
		commentLength := binary.LittleEndian.Uint32(m.vorbisComment[offset : offset+4])
		offset += 4

		if offset+int(commentLength) > len(m.vorbisComment) {
			return errors.New("invalid Vorbis comment block")
		}
		comment := string(m.vorbisComment[offset : offset+int(commentLength)])
		m.tags = append(m.tags, comment)
		offset += int(commentLength)
	}
	return nil
}

// Parse a picture block and extract its specification and data.
func (m *Metaflac) parsePictureBlock(picture []byte) {
	var offset int
	if len(picture) < 32 {
		return // Not enough data to parse
	}
	spec := PictureSpec{}
	spec.Type = binary.BigEndian.Uint32(picture[offset:])
	offset += 4
	mimeLength := binary.BigEndian.Uint32(picture[offset:])
	offset += 4
	spec.Mime = string(picture[offset : offset+int(mimeLength)])
	offset += int(mimeLength)
	descriptionLength := binary.BigEndian.Uint32(picture[offset:])
	offset += 4
	spec.Description = string(picture[offset : offset+int(descriptionLength)])
	offset += int(descriptionLength)
	spec.Width = binary.BigEndian.Uint32(picture[offset:])
	offset += 4
	spec.Height = binary.BigEndian.Uint32(picture[offset:])
	offset += 4
	spec.Depth = binary.BigEndian.Uint32(picture[offset:])
	offset += 4
	spec.Colors = binary.BigEndian.Uint32(picture[offset:])
	offset += 4
	pictureDataLength := binary.BigEndian.Uint32(picture[offset:])
	offset += 4
	if offset+int(pictureDataLength) > len(picture) {
		return // Not enough data
	}
	pictureData := picture[offset : offset+int(pictureDataLength)]
	m.picturesDatas = append(m.picturesDatas, pictureData)
	m.picturesSpecs = append(m.picturesSpecs, spec)
}

// GetPicturesSpecs returns the specifications of all imported pictures.
func (m *Metaflac) GetPicturesSpecs() []PictureSpec {
	return m.picturesSpecs
}

// GetMd5sum returns the MD5 signature from the STREAMINFO block.
func (m *Metaflac) GetMd5sum() string {
	if len(m.streamInfo) < 34 {
		return ""
	}
	return hex.EncodeToString(m.streamInfo[18:34])
}

// GetMinBlocksize returns the minimum block size from the STREAMINFO block.
func (m *Metaflac) GetMinBlocksize() uint16 {
	if len(m.streamInfo) < 2 {
		return 0
	}
	return binary.BigEndian.Uint16(m.streamInfo[0:2])
}

// GetMaxBlocksize returns the maximum block size from the STREAMINFO block.
func (m *Metaflac) GetMaxBlocksize() uint16 {
	if len(m.streamInfo) < 4 {
		return 0
	}
	return binary.BigEndian.Uint16(m.streamInfo[2:4])
}

// GetMinFramesize returns the minimum frame size from the STREAMINFO block.
func (m *Metaflac) GetMinFramesize() uint32 {
	if len(m.streamInfo) < 7 {
		return 0
	}
	return uint32(m.streamInfo[4])<<16 | uint32(m.streamInfo[5])<<8 | uint32(m.streamInfo[6])
}

// GetMaxFramesize returns the maximum frame size from the STREAMINFO block.
func (m *Metaflac) GetMaxFramesize() uint32 {
	if len(m.streamInfo) < 10 {
		return 0
	}
	return uint32(m.streamInfo[7])<<16 | uint32(m.streamInfo[8])<<8 | uint32(m.streamInfo[9])
}

// GetSampleRate returns the sample rate from the STREAMINFO block.
func (m *Metaflac) GetSampleRate() uint32 {
	if len(m.streamInfo) < 18 {
		return 0
	}
	return (uint32(m.streamInfo[10])<<12 | uint32(m.streamInfo[11])<<4 | uint32(m.streamInfo[12])>>4)
}

// GetChannels returns the number of channels from the STREAMINFO block.
func (m *Metaflac) GetChannels() uint8 {
	if len(m.streamInfo) < 13 {
		return 0
	}
	return ((m.streamInfo[12] & 0x0E) >> 1) + 1
}

// GetBps returns the bits per sample from the STREAMINFO block.
func (m *Metaflac) GetBps() uint8 {
	if len(m.streamInfo) < 14 {
		return 0
	}
	return (((m.streamInfo[12] & 0x01) << 4) | (m.streamInfo[13] >> 4)) + 1
}

// GetTotalSamples returns the total number of samples from the STREAMINFO block.
func (m *Metaflac) GetTotalSamples() uint64 {
	if len(m.streamInfo) < 18 {
		return 0
	}
	return uint64(m.streamInfo[13]&0x0F)<<32 |
		uint64(m.streamInfo[14])<<24 |
		uint64(m.streamInfo[15])<<16 |
		uint64(m.streamInfo[16])<<8 |
		uint64(m.streamInfo[17])
}

// GetVendorTag returns the vendor string from the VORBIS_COMMENT block.
func (m *Metaflac) GetVendorTag() string {
	return m.vendorString
}

// GetTag retrieves all tags matching the given name.
func (m *Metaflac) GetTag(name string) []string {
	var matchingTags []string
	for _, tag := range m.tags {
		if strings.HasPrefix(tag, name+"=") {
			matchingTags = append(matchingTags, tag)
		}
	}
	return matchingTags
}

// RemoveTag removes all tags with the given name.
func (m *Metaflac) RemoveTag(name string) {
	var filteredTags []string
	for _, tag := range m.tags {
		if !strings.HasPrefix(tag, name+"=") {
			filteredTags = append(filteredTags, tag)
		}
	}
	m.tags = filteredTags
}

// RemoveFirstTag removes the first tag with the given name.
func (m *Metaflac) RemoveFirstTag(name string) {
	for i, tag := range m.tags {
		if strings.HasPrefix(tag, name+"=") {
			m.tags = append(m.tags[:i], m.tags[i+1:]...)
			break
		}
	}
}

// RemoveAllTags removes all tags, leaving only the vendor string.
func (m *Metaflac) RemoveAllTags() {
	m.tags = []string{}
}

// SetTag adds a new tag.
func (m *Metaflac) SetTag(field string) error {
	if !strings.Contains(field, "=") {
		return errors.New("malformed Vorbis comment field; missing '=' character")
	}
	m.tags = append(m.tags, field)
	return nil
}

// ImportPicture imports a picture into a PICTURE metadata block.
func (m *Metaflac) ImportPicture(pictureData []byte, spec PictureSpec) {
	pictureBlock := m.buildPictureBlock(pictureData, spec)
	m.pictures = append(m.pictures, pictureBlock)
	m.picturesSpecs = append(m.picturesSpecs, spec)
}

// GetAllTags returns all tags.
func (m *Metaflac) GetAllTags() []string {
	return m.tags
}

// buildPictureBlock builds a picture block.
func (m *Metaflac) buildPictureBlock(picture []byte, spec PictureSpec) []byte {
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
	binary.Write(&buffer, binary.BigEndian, uint32(len(picture)))
	buffer.Write(picture)

	return buffer.Bytes()
}

// buildMetadataBlock builds a metadata block.
func (m *Metaflac) buildMetadataBlock(blockType int, block []byte, isLast bool) []byte {
	var header bytes.Buffer

	if isLast {
		blockType |= 0x80
	}
	header.WriteByte(byte(blockType))

	lengthBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthBytes, uint32(len(block)))
	header.Write(lengthBytes[1:])

	return append(header.Bytes(), block...)
}

// buildMetadata builds all metadata blocks.
func (m *Metaflac) buildMetadata() [][]byte {
	var metadata [][]byte

	metadata = append(metadata, m.buildMetadataBlock(STREAMINFO, m.streamInfo, false))

	for _, block := range m.blocks {
		metadata = append(metadata, m.buildMetadataBlock(block.BlockType, block.Data, false))
	}

	vorbisCommentBlock := formatVorbisComment(m.vendorString, m.tags)
	metadata = append(metadata, m.buildMetadataBlock(VORBIS_COMMENT, vorbisCommentBlock, false))

	for _, picture := range m.pictures {
		metadata = append(metadata, m.buildMetadataBlock(PICTURE, picture, false))
	}

	if m.padding == nil {
		m.padding = make([]byte, 16384)
	}
	metadata = append(metadata, m.buildMetadataBlock(PADDING, m.padding, true))

	return metadata
}

// buildStream rebuilds the FLAC stream with updated metadata.
func (m *Metaflac) buildStream() []byte {
	metadata := m.buildMetadata()
	var buffer bytes.Buffer

	buffer.Write(m.buffer[:4]) // Write the 'fLaC' marker
	for _, block := range metadata {
		buffer.Write(block)
	}
	buffer.Write(m.buffer[m.framesOffset:])

	return buffer.Bytes()
}

// GetBuffer returns the modified FLAC buffer.
func (m *Metaflac) GetBuffer() []byte {
	return m.buildStream()
}
