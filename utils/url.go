package utils

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/d-fi/GoFi/logger"
)

// CheckURLFileSize performs a HEAD request to check the availability of a URL
// and returns the content length if available.
func CheckURLFileSize(url string) (int, error) {
	logger.Debug("Checking URL file size for: %s", url)

	resp, err := http.Head(url)
	if err != nil {
		logger.Debug("Error during HEAD request: %v", err)
		return 0, err
	}
	defer resp.Body.Close()

	contentLength := resp.Header.Get("Content-Length")
	if contentLength != "" {
		var size int
		fmt.Sscanf(contentLength, "%d", &size)
		logger.Debug("Determined file size: %d bytes", size)
		return size, nil
	}

	logger.Debug("Unable to determine file size for URL: %s", url)
	return 0, errors.New("unable to determine file size")
}
