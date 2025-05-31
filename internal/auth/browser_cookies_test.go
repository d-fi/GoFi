package auth

import (
	"runtime"
	"strings"
	"testing"
)

func TestNewCookieReader(t *testing.T) {
	tests := []struct {
		name    string
		browser BrowserType
		want    BrowserType
	}{
		{"Chrome", Chrome, Chrome},
		{"Firefox", Firefox, Firefox},
		{"Safari", Safari, Safari},
		{"Edge", Edge, Edge},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cr := NewCookieReader(tt.browser)
			if cr.browser != tt.want {
				t.Errorf("NewCookieReader() browser = %v, want %v", cr.browser, tt.want)
			}
			if cr.os != runtime.GOOS {
				t.Errorf("NewCookieReader() os = %v, want %v", cr.os, runtime.GOOS)
			}
		})
	}
}

func TestParseCookieString(t *testing.T) {
	tests := []struct {
		name        string
		cookieString string
		want        string
		wantErr     bool
	}{
		{
			name:        "Standard cookie format",
			cookieString: "sid=abc123; arl=xyz789; uid=456",
			want:        "xyz789",
			wantErr:     false,
		},
		{
			name:        "ARL only",
			cookieString: "arl=myarlvalue123",
			want:        "myarlvalue123",
			wantErr:     false,
		},
		{
			name:        "Raw ARL value",
			cookieString: "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0u1v2w3x4y5z6a7b8c9d0e1f2g3h4i5j6k7l8m9n0o1p2q3r4s5t6u7v8w9x0y1z2",
			want:        "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0u1v2w3x4y5z6a7b8c9d0e1f2g3h4i5j6k7l8m9n0o1p2q3r4s5t6u7v8w9x0y1z2",
			wantErr:     false,
		},
		{
			name:        "No ARL cookie",
			cookieString: "sid=abc123; uid=456",
			want:        "",
			wantErr:     true,
		},
		{
			name:        "Empty string",
			cookieString: "",
			want:        "",
			wantErr:     true,
		},
		{
			name:        "Case insensitive ARL",
			cookieString: "ARL=upperCaseValue",
			want:        "upperCaseValue",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseCookieString(tt.cookieString)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCookieString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseCookieString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetCookiePath(t *testing.T) {
	// This test checks that getCookiePath returns valid paths for different browsers
	browsers := []BrowserType{Chrome, Firefox, Edge}
	if runtime.GOOS == "darwin" {
		browsers = append(browsers, Safari)
	}

	for _, browser := range browsers {
		t.Run(string(browser), func(t *testing.T) {
			cr := NewCookieReader(browser)
			path, err := cr.getCookiePath()
			
			// We expect either a valid path or a "file not found" error
			if err != nil {
				// Check if it's a "file not found" error (expected) or another error (unexpected)
				if !strings.Contains(err.Error(), "cookie file not found") && 
				   !strings.Contains(err.Error(), "not found") &&
				   !strings.Contains(err.Error(), "no Firefox profile found") {
					t.Errorf("getCookiePath() returned unexpected error for %s: %v", browser, err)
				}
			} else {
				// If no error, path should not be empty
				if path == "" {
					t.Errorf("getCookiePath() returned empty path for %s", browser)
				}
			}
		})
	}
}

func TestDecryptAES128CBC(t *testing.T) {
	// Test basic AES-128-CBC decryption functionality
	cr := NewCookieReader(Chrome)
	
	// This is a basic test with known values
	// In real scenarios, the encrypted data would come from the browser
	key := []byte("0123456789abcdef") // 16 bytes for AES-128
	
	// Test with empty data
	_, err := cr.decryptAES128CBC(key, []byte{})
	if err == nil {
		t.Error("decryptAES128CBC() should fail with empty data")
	}
	
	// Test with data shorter than block size
	_, err = cr.decryptAES128CBC(key, []byte("short"))
	if err == nil {
		t.Error("decryptAES128CBC() should fail with data shorter than block size")
	}
}