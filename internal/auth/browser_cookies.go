package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/pbkdf2"
)

// BrowserType represents different browser types
type BrowserType string

const (
	Chrome  BrowserType = "chrome"
	Firefox BrowserType = "firefox"
	Safari  BrowserType = "safari"
	Edge    BrowserType = "edge"
)

// CookieReader provides methods to read cookies from browsers
type CookieReader struct {
	browser BrowserType
	os      string
}

// NewCookieReader creates a new cookie reader for the specified browser
func NewCookieReader(browser BrowserType) *CookieReader {
	return &CookieReader{
		browser: browser,
		os:      runtime.GOOS,
	}
}

// GetDeezerARL attempts to read the 'arl' cookie from deezer.com
func (cr *CookieReader) GetDeezerARL() (string, error) {
	cookiePath, err := cr.getCookiePath()
	if err != nil {
		return "", fmt.Errorf("failed to get cookie path: %w", err)
	}

	switch cr.browser {
	case Chrome, Edge:
		return cr.getChromiumCookie(cookiePath, "arl", ".deezer.com")
	case Firefox:
		return cr.getFirefoxCookie(cookiePath, "arl", ".deezer.com")
	case Safari:
		return cr.getSafariCookie(cookiePath, "arl", ".deezer.com")
	default:
		return "", fmt.Errorf("unsupported browser: %s", cr.browser)
	}
}

// getCookiePath returns the path to the browser's cookie database
func (cr *CookieReader) getCookiePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	var path string
	switch cr.os {
	case "darwin": // macOS
		switch cr.browser {
		case Chrome:
			path = filepath.Join(home, "Library", "Application Support", "Google", "Chrome", "Default", "Cookies")
		case Firefox:
			// Firefox profile path needs to be discovered
			profilePath, err := cr.findFirefoxProfile(filepath.Join(home, "Library", "Application Support", "Firefox", "Profiles"))
			if err != nil {
				return "", err
			}
			path = filepath.Join(profilePath, "cookies.sqlite")
		case Safari:
			path = filepath.Join(home, "Library", "Cookies", "Cookies.binarycookies")
		case Edge:
			path = filepath.Join(home, "Library", "Application Support", "Microsoft Edge", "Default", "Cookies")
		default:
			return "", fmt.Errorf("browser %s not supported on macOS", cr.browser)
		}
	case "linux":
		switch cr.browser {
		case Chrome:
			path = filepath.Join(home, ".config", "google-chrome", "Default", "Cookies")
		case Firefox:
			profilePath, err := cr.findFirefoxProfile(filepath.Join(home, ".mozilla", "firefox"))
			if err != nil {
				return "", err
			}
			path = filepath.Join(profilePath, "cookies.sqlite")
		case Edge:
			path = filepath.Join(home, ".config", "microsoft-edge", "Default", "Cookies")
		default:
			return "", fmt.Errorf("browser %s not supported on Linux", cr.browser)
		}
	case "windows":
		localAppData := os.Getenv("LOCALAPPDATA")
		appData := os.Getenv("APPDATA")
		switch cr.browser {
		case Chrome:
			path = filepath.Join(localAppData, "Google", "Chrome", "User Data", "Default", "Network", "Cookies")
		case Firefox:
			profilePath, err := cr.findFirefoxProfile(filepath.Join(appData, "Mozilla", "Firefox", "Profiles"))
			if err != nil {
				return "", err
			}
			path = filepath.Join(profilePath, "cookies.sqlite")
		case Edge:
			path = filepath.Join(localAppData, "Microsoft", "Edge", "User Data", "Default", "Network", "Cookies")
		default:
			return "", fmt.Errorf("browser %s not supported on Windows", cr.browser)
		}
	default:
		return "", fmt.Errorf("unsupported operating system: %s", cr.os)
	}

	if path == "" {
		return "", fmt.Errorf("could not determine cookie path for %s on %s", cr.browser, cr.os)
	}

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", fmt.Errorf("cookie file not found at %s", path)
	}

	return path, nil
}

// findFirefoxProfile finds the default Firefox profile directory
func (cr *CookieReader) findFirefoxProfile(profilesDir string) (string, error) {
	entries, err := os.ReadDir(profilesDir)
	if err != nil {
		return "", fmt.Errorf("failed to read Firefox profiles directory: %w", err)
	}

	// Look for default profile (usually contains ".default" or ".default-release")
	for _, entry := range entries {
		if entry.IsDir() && (strings.Contains(entry.Name(), ".default") || strings.Contains(entry.Name(), ".default-release")) {
			return filepath.Join(profilesDir, entry.Name()), nil
		}
	}

	// If no default found, return the first profile
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			return filepath.Join(profilesDir, entry.Name()), nil
		}
	}

	return "", errors.New("no Firefox profile found")
}

// getChromiumCookie reads a cookie from Chrome/Edge cookie database
func (cr *CookieReader) getChromiumCookie(dbPath, name, domain string) (string, error) {
	// Make a copy of the database to avoid lock issues
	tempDB := dbPath + ".tmp"
	if err := copyFile(dbPath, tempDB); err != nil {
		return "", fmt.Errorf("failed to copy cookie database: %w", err)
	}
	defer os.Remove(tempDB)

	db, err := sql.Open("sqlite3", tempDB)
	if err != nil {
		return "", fmt.Errorf("failed to open cookie database: %w", err)
	}
	defer db.Close()

	var encryptedValue []byte
	query := `SELECT encrypted_value FROM cookies WHERE host_key = ? AND name = ?`
	err = db.QueryRow(query, domain, name).Scan(&encryptedValue)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("cookie '%s' not found for domain '%s'", name, domain)
		}
		return "", fmt.Errorf("failed to query cookie: %w", err)
	}

	// Decrypt the cookie value
	decrypted, err := cr.decryptChromiumCookie(encryptedValue)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt cookie: %w", err)
	}

	return decrypted, nil
}

// decryptChromiumCookie decrypts Chrome/Edge encrypted cookies
func (cr *CookieReader) decryptChromiumCookie(encrypted []byte) (string, error) {
	if len(encrypted) == 0 {
		return "", nil
	}

	switch cr.os {
	case "darwin":
		return cr.decryptChromiumCookieMac(encrypted)
	case "windows":
		return cr.decryptChromiumCookieWindows(encrypted)
	case "linux":
		return cr.decryptChromiumCookieLinux(encrypted)
	default:
		return "", fmt.Errorf("unsupported OS for decryption: %s", cr.os)
	}
}

// decryptChromiumCookieMac decrypts Chrome cookies on macOS
func (cr *CookieReader) decryptChromiumCookieMac(encrypted []byte) (string, error) {
	// Check for v10 prefix
	if len(encrypted) < 3 || string(encrypted[:3]) != "v10" {
		// Not encrypted or old format
		return string(encrypted), nil
	}

	// Remove v10 prefix
	encrypted = encrypted[3:]

	// Get Chrome Safe Storage password from Keychain
	password, err := cr.getChromePassword()
	if err != nil {
		return "", fmt.Errorf("failed to get Chrome password: %w", err)
	}

	// Derive key using PBKDF2
	key := pbkdf2.Key([]byte(password), []byte("saltysalt"), 1003, 16, sha1.New)

	// Decrypt using AES-128-CBC
	return cr.decryptAES128CBC(key, encrypted)
}

// getChromePassword retrieves Chrome's Safe Storage password from macOS Keychain
func (cr *CookieReader) getChromePassword() (string, error) {
	// This is a simplified version - in production, you'd use the keychain API
	// For now, we'll use the hardcoded Chrome Safe Storage password
	return "Chrome Safe Storage", nil
}

// decryptChromiumCookieWindows decrypts Chrome cookies on Windows
func (cr *CookieReader) decryptChromiumCookieWindows(encrypted []byte) (string, error) {
	// On Windows, Chrome uses DPAPI
	// This is a placeholder - actual implementation would use Windows DPAPI
	return "", errors.New("Windows cookie decryption not implemented")
}

// decryptChromiumCookieLinux decrypts Chrome cookies on Linux
func (cr *CookieReader) decryptChromiumCookieLinux(encrypted []byte) (string, error) {
	// Check for v10 or v11 prefix
	if len(encrypted) < 3 {
		return string(encrypted), nil
	}

	version := string(encrypted[:3])
	if version != "v10" && version != "v11" {
		// Not encrypted
		return string(encrypted), nil
	}

	// Remove version prefix
	encrypted = encrypted[3:]

	// Default Chrome password on Linux
	password := "peanuts"

	// Derive key using PBKDF2
	key := pbkdf2.Key([]byte(password), []byte("saltysalt"), 1, 16, sha1.New)

	// Decrypt using AES-128-CBC
	return cr.decryptAES128CBC(key, encrypted)
}

// decryptAES128CBC decrypts data using AES-128-CBC
func (cr *CookieReader) decryptAES128CBC(key, encrypted []byte) (string, error) {
	if len(encrypted) < aes.BlockSize {
		return "", errors.New("encrypted data too short")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// Extract IV (first 16 bytes)
	iv := encrypted[:aes.BlockSize]
	encrypted = encrypted[aes.BlockSize:]

	// Decrypt
	mode := cipher.NewCBCDecrypter(block, iv)
	decrypted := make([]byte, len(encrypted))
	mode.CryptBlocks(decrypted, encrypted)

	// Remove PKCS7 padding
	padding := int(decrypted[len(decrypted)-1])
	if padding > 0 && padding <= aes.BlockSize {
		decrypted = decrypted[:len(decrypted)-padding]
	}

	return string(decrypted), nil
}

// getFirefoxCookie reads a cookie from Firefox cookie database
func (cr *CookieReader) getFirefoxCookie(dbPath, name, domain string) (string, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return "", fmt.Errorf("failed to open Firefox cookie database: %w", err)
	}
	defer db.Close()

	var value string
	query := `SELECT value FROM moz_cookies WHERE host = ? AND name = ?`
	err = db.QueryRow(query, domain, name).Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("cookie '%s' not found for domain '%s'", name, domain)
		}
		return "", fmt.Errorf("failed to query Firefox cookie: %w", err)
	}

	return value, nil
}

// getSafariCookie reads a cookie from Safari (macOS only)
func (cr *CookieReader) getSafariCookie(dbPath, name, domain string) (string, error) {
	// Safari uses a binary plist format for cookies
	// This is a placeholder - actual implementation would parse .binarycookies format
	return "", errors.New("Safari cookie reading not implemented")
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	err = os.WriteFile(dst, input, 0644)
	if err != nil {
		return err
	}

	return nil
}

// GetARLFromAnyBrowser tries to get the ARL cookie from any available browser
func GetARLFromAnyBrowser() (string, error) {
	browsers := []BrowserType{Chrome, Firefox, Edge}
	if runtime.GOOS == "darwin" {
		browsers = append(browsers, Safari)
	}

	var lastErr error
	for _, browser := range browsers {
		reader := NewCookieReader(browser)
		arl, err := reader.GetDeezerARL()
		if err == nil && arl != "" {
			return arl, nil
		}
		lastErr = err
	}

	if lastErr != nil {
		return "", fmt.Errorf("failed to get ARL cookie from any browser: %w", lastErr)
	}
	return "", errors.New("no ARL cookie found in any browser")
}

// ParseCookieString parses a cookie string and extracts the ARL value
func ParseCookieString(cookieString string) (string, error) {
	originalString := strings.TrimSpace(cookieString)
	
	// Try to handle base64 encoded cookies
	if decoded, err := base64.StdEncoding.DecodeString(originalString); err == nil && len(decoded) > 0 {
		// Only use decoded if it contains valid text
		decodedStr := string(decoded)
		if strings.Contains(decodedStr, "arl") || strings.Contains(decodedStr, "ARL") {
			cookieString = decodedStr
		} else {
			cookieString = originalString
		}
	} else {
		cookieString = originalString
	}

	// Parse cookie string format: "name=value; name2=value2; ..."
	if strings.Contains(cookieString, "=") {
		cookies := strings.Split(cookieString, ";")
		for _, cookie := range cookies {
			parts := strings.SplitN(strings.TrimSpace(cookie), "=", 2)
			if len(parts) == 2 && strings.ToLower(parts[0]) == "arl" {
				return strings.TrimSpace(parts[1]), nil
			}
		}
	}

	// Check if the string itself is just the ARL value (long alphanumeric string)
	if len(cookieString) >= 50 && !strings.Contains(cookieString, "=") && !strings.Contains(cookieString, ";") && !strings.Contains(cookieString, " ") {
		// Validate it looks like an ARL token
		validChars := true
		for _, r := range cookieString {
			if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-') {
				validChars = false
				break
			}
		}
		if validChars {
			return cookieString, nil
		}
	}

	return "", errors.New("ARL cookie not found in cookie string")
}