package decrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/hex"
	"fmt"

	"github.com/d-fi/GoFi/logger"
	"golang.org/x/crypto/blowfish"
)

type TrackType struct {
	MD5_ORIGIN    string
	SNG_ID        string
	MEDIA_VERSION string
}

// Md5Hash generates an MD5 hash for the given string data.
func Md5Hash(data string) string {
	logger.Debug("Generating MD5 hash for data: %s", data)
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}

// GetSongFileName generates the song file name using encryption.
func GetSongFileName(track *TrackType, quality int) string {
	step1 := fmt.Sprintf("%s¤%d¤%s¤%s", track.MD5_ORIGIN, quality, track.SNG_ID, track.MEDIA_VERSION)
	logger.Debug("Step 1 - Combined track details: %s", step1)

	step2 := Md5Hash(step1) + "¤" + step1 + "¤"
	for len(step2)%16 > 0 {
		step2 += " "
	}
	logger.Debug("Step 2 - MD5 hash with padding: %s", step2)

	cipherKey := []byte("jo6aey6haid2Teih")
	block, err := aes.NewCipher(cipherKey)
	if err != nil {
		logger.Error("Failed to create AES cipher: %v", err)
		panic(err)
	}

	encrypted := make([]byte, len(step2))
	encryptECB(block, []byte(step2), encrypted)
	logger.Debug("Encrypted song file name: %s", hex.EncodeToString(encrypted))

	return hex.EncodeToString(encrypted)
}

// ECB mode encryption helper function since Go's crypto library doesn't directly support ECB.
func encryptECB(block cipher.Block, src, dst []byte) {
	bs := block.BlockSize()
	if len(src)%bs != 0 {
		logger.Error("Plaintext is not a multiple of the block size")
		panic("plaintext is not a multiple of the block size")
	}

	for len(src) > 0 {
		block.Encrypt(dst, src[:bs])
		src = src[bs:]
		dst = dst[bs:]
	}
}

// GetBlowfishKey generates a blowfish key using the track ID.
func GetBlowfishKey(trackID string) string {
	SECRET := "g4el58wc" + "0zvf9na1"
	idMd5 := Md5Hash(trackID)
	bfKey := ""
	for i := 0; i < 16; i++ {
		bfKey += string(rune(idMd5[i]) ^ rune(idMd5[i+16]) ^ rune(SECRET[i]))
	}
	logger.Debug("Generated blowfish key: %s", bfKey)
	return bfKey
}

// DecryptChunk decrypts a chunk of data using the blowfish key.
func DecryptChunk(chunk []byte, blowfishKey string) []byte {
	iv := []byte{0, 1, 2, 3, 4, 5, 6, 7}
	block, err := blowfish.NewCipher([]byte(blowfishKey))
	if err != nil {
		logger.Error("Failed to create blowfish cipher: %v", err)
		panic(err)
	}

	dst := make([]byte, len(chunk))
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(dst, chunk)
	logger.Debug("Decrypted chunk: %x", dst)
	return dst
}

// DecryptDownload decrypts the downloaded track using the blowfish key.
func DecryptDownload(source []byte, trackID string) []byte {
	chunkSize := 2048
	blowfishKey := GetBlowfishKey(trackID)
	logger.Debug("Decrypting download with track ID: %s", trackID)

	i := 0
	position := 0
	destBuffer := make([]byte, len(source))

	for i := 0; i < len(destBuffer); i++ {
		destBuffer[i] = 0
	}

	for position < len(source) {
		size := len(source) - position
		chunkSize = 2048
		if size < 2048 {
			chunkSize = size
		}

		chunk := make([]byte, chunkSize)
		copy(chunk, source[position:position+chunkSize])

		if i%3 > 0 || chunkSize < 2048 {
			copy(destBuffer[position:], chunk)
		} else {
			decryptedChunk := DecryptChunk(chunk, blowfishKey)
			copy(destBuffer[position:], decryptedChunk)
		}

		position += chunkSize
		i++
	}

	logger.Debug("Decryption completed for track ID: %s", trackID)
	return destBuffer
}
