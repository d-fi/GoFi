package utils

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/d-fi/GoFi/logger"
)

// CheckURLFileSize performs a HEAD request to check the availability of a URL
// and returns the content length if available.
func CheckURLFileSize(url string) (int64, error) {
	logger.Debug("Checking URL file size for: %s", url)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Head(url)
	if err != nil {
		logger.Debug("Error during HEAD request: %v", err)
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		logger.Debug("Non-success status code: %d", resp.StatusCode)
		return 0, fmt.Errorf("received non-success status code: %d", resp.StatusCode)
	}

	contentLength := resp.ContentLength
	if contentLength >= 0 {
		logger.Debug("Determined file size: %d bytes", contentLength)
		return contentLength, nil
	}

	logger.Debug("Content-Length header is missing")
	return 0, errors.New("content-length header is missing")
}
