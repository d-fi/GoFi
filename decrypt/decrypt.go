package decrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/blowfish"
)

type TrackType struct {
	MD5_ORIGIN    string
	SNG_ID        string
	MEDIA_VERSION string
}

func Md5Hash(data string) string {
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}

func GetSongFileName(track *TrackType, quality int) string {
	// Step 1: Combine track details into a specific format
	step1 := fmt.Sprintf("%s¤%d¤%s¤%s", track.MD5_ORIGIN, quality, track.SNG_ID, track.MEDIA_VERSION)

	// Step 2: Generate MD5 hash and concatenate it with step1, ensuring it's a multiple of 16 bytes
	step2 := Md5Hash(step1) + "¤" + step1 + "¤"
	for len(step2)%16 > 0 {
		step2 += " "
	}

	// AES-128-ECB encryption
	cipherKey := []byte("jo6aey6haid2Teih")
	block, err := aes.NewCipher(cipherKey)
	if err != nil {
		panic(err)
	}

	// Encrypt in ECB mode manually since Go's AES doesn't support ECB directly
	encrypted := make([]byte, len(step2))
	encryptECB(block, []byte(step2), encrypted)

	return hex.EncodeToString(encrypted)
}

// ECB mode encryption helper function since Go's crypto library doesn't directly support ECB.
func encryptECB(block cipher.Block, src, dst []byte) {
	bs := block.BlockSize()
	if len(src)%bs != 0 {
		panic("plaintext is not a multiple of the block size")
	}

	for len(src) > 0 {
		block.Encrypt(dst, src[:bs])
		src = src[bs:]
		dst = dst[bs:]
	}
}

func GetBlowfishKey(trackID string) string {
	SECRET := "g4el58wc" + "0zvf9na1"
	idMd5 := Md5Hash(trackID)
	bfKey := ""
	for i := 0; i < 16; i++ {
		bfKey += string(rune(idMd5[i]) ^ rune(idMd5[i+16]) ^ rune(SECRET[i]))
	}
	return bfKey
}

func DecryptChunk(chunk []byte, blowfishKey string) []byte {
	iv := []byte{0, 1, 2, 3, 4, 5, 6, 7}
	block, err := blowfish.NewCipher([]byte(blowfishKey))
	if err != nil {
		panic(err)
	}

	dst := make([]byte, len(chunk))
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(dst, chunk)
	return dst
}

func DecryptDownload(source []byte, trackID string) []byte {
	chunkSize := 2048
	blowfishKey := GetBlowfishKey(trackID)
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

	return destBuffer
}
