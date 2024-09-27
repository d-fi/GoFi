package metaflac

import (
	"bytes"
	"encoding/binary"
)

// formatVorbisComment formats the Vorbis comment block.
func formatVorbisComment(vendorString string, commentList []string) []byte {
	var buffer bytes.Buffer

	binary.Write(&buffer, binary.LittleEndian, uint32(len(vendorString)))
	buffer.WriteString(vendorString)

	binary.Write(&buffer, binary.LittleEndian, uint32(len(commentList)))
	for _, comment := range commentList {
		binary.Write(&buffer, binary.LittleEndian, uint32(len(comment)))
		buffer.WriteString(comment)
	}

	return buffer.Bytes()
}

// Helper function to generate a minimal FLAC buffer.
func generateMinimalFlac() []byte {
	var buffer bytes.Buffer

	// Write the 'fLaC' marker
	buffer.WriteString("fLaC")

	// Build a minimal STREAMINFO block (block type 0, length 34 bytes)
	streamInfo := make([]byte, 34)
	// Set some default values in STREAMINFO
	binary.BigEndian.PutUint16(streamInfo[0:], 4096) // min block size
	binary.BigEndian.PutUint16(streamInfo[2:], 4096) // max block size
	streamInfo[18] = 0x12                            // MD5 placeholder
	streamInfo[19] = 0x34                            // MD5 placeholder

	// Build the STREAMINFO metadata block
	streamInfoBlock := buildTestMetadataBlock(STREAMINFO, streamInfo, false)
	buffer.Write(streamInfoBlock)

	// Add a PADDING block as the last block
	padding := make([]byte, 0)
	paddingBlock := buildTestMetadataBlock(PADDING, padding, true)
	buffer.Write(paddingBlock)

	// Add dummy frame data
	buffer.WriteString("dummy frame data")

	return buffer.Bytes()
}

// Helper function to build a metadata block for testing.
func buildTestMetadataBlock(blockType int, block []byte, isLast bool) []byte {
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

// Helper function to compare two slices of strings
func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Helper function to compare two PictureSpec structs
func comparePictureSpec(a, b PictureSpec) bool {
	return a.Type == b.Type &&
		a.Mime == b.Mime &&
		a.Description == b.Description &&
		a.Width == b.Width &&
		a.Height == b.Height &&
		a.Depth == b.Depth &&
		a.Colors == b.Colors
}
