package utils

import (
	"errors"
	"fmt"
	"net/http"
)

// CheckURLFileSize performs a HEAD request to check the availability of a URL
// and returns the content length if available.
func CheckURLFileSize(url string) (int, error) {
	resp, err := http.Head(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if contentLength := resp.Header.Get("Content-Length"); contentLength != "" {
		var size int
		fmt.Sscanf(contentLength, "%d", &size)
		return size, nil
	}
	return 0, errors.New("unable to determine file size")
}
