package decrypt

import (
	"crypto/aes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
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
	step1 := fmt.Sprintf("%s¤%d¤%s¤%s", track.MD5_ORIGIN, quality, track.SNG_ID, track.MEDIA_VERSION)

	step2 := Md5Hash(step1) + "¤" + step1 + "¤"
	for len(step2)%16 > 0 {
		step2 += " "
	}

	cipherKey := []byte("jo6aey6haid2Teih")
	block, err := aes.NewCipher(cipherKey)
	if err != nil {
		panic(err)
	}

	cipherText := make([]byte, len(step2))
	block.Encrypt(cipherText, []byte(step2))

	return hex.EncodeToString(cipherText)
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
	block, err := aes.NewCipher([]byte(blowfishKey))
	if err != nil {
		panic(err)
	}

	cipherText := make([]byte, len(chunk))
	block.Decrypt(cipherText, chunk)

	return cipherText
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

		var chunkString string
		if i%3 > 0 || chunkSize < 2048 {
			chunkString = string(chunk)
		} else {
			chunkString = string(DecryptChunk(chunk, blowfishKey))
		}

		copy(destBuffer[position:position+len(chunkString)], chunkString)
		position += chunkSize
		i++
	}

	return destBuffer
}
