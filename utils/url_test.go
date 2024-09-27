package utils

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCheckURLFileSize(t *testing.T) {
	tests := []struct {
		name           string
		handler        http.HandlerFunc
		expectedSize   int64
		expectedErrMsg string
		timeout        *time.Duration
	}{
		{
			name: "Valid Content-Length",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Length", "12345")
				w.WriteHeader(http.StatusOK)
			},
			expectedSize: 12345,
			timeout:      nil, // Use default timeout
		},
		{
			name: "Missing Content-Length",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			expectedErrMsg: "content-length header is missing",
			timeout:        nil, // Use default timeout
		},
		{
			name: "Non-success status code",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			},
			expectedErrMsg: "received non-success status code: 404",
			timeout:        nil, // Use default timeout
		},
		{
			name: "Timeout",
			handler: func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(2 * time.Second) // Sleep longer than client timeout
			},
			expectedErrMsg: "Client.Timeout exceeded",
			timeout:        func() *time.Duration { t := 1 * time.Second; return &t }(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(test.handler)
			defer server.Close()

			size, err := CheckURLFileSize(server.URL, test.timeout)

			if test.expectedErrMsg != "" {
				assert.Error(t, err, "Expected an error for test '%s'", test.name)
				assert.Contains(t, err.Error(), test.expectedErrMsg, "Error message mismatch for test '%s'", test.name)
			} else {
				assert.NoError(t, err, "Unexpected error for test '%s'", test.name)
				assert.Equal(t, test.expectedSize, size, "File size mismatch for test '%s'", test.name)
			}
		})
	}
}
