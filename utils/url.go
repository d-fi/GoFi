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
// The timeout parameter is optional; if nil, it defaults to 10 seconds.
func CheckURLFileSize(url string, timeout *time.Duration) (int64, error) {
	var clientTimeout time.Duration

	if timeout != nil && *timeout > 0 {
		clientTimeout = *timeout
	} else {
		clientTimeout = 10 * time.Second
	}

	logger.Debug("Checking URL file size for: %s with timeout: %s", url, clientTimeout)

	client := &http.Client{
		Timeout: clientTimeout,
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
