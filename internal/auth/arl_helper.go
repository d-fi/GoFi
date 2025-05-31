package auth

import (
	"fmt"
	"os"
	"strings"
)

// GetARLToken attempts to get the ARL token from various sources
// Priority: 1. Environment variable, 2. Browser cookies, 3. Config file
func GetARLToken() (string, error) {
	// First, check environment variable
	if arl := os.Getenv("DEEZER_ARL"); arl != "" {
		return arl, nil
	}

	// Try to get from browser cookies
	arl, err := GetARLFromAnyBrowser()
	if err == nil && arl != "" {
		return arl, nil
	}

	// Return error if no ARL found
	return "", fmt.Errorf("ARL token not found in environment or browser cookies: %w", err)
}

// ValidateARLToken performs basic validation on an ARL token
func ValidateARLToken(arl string) error {
	// Remove any whitespace
	arl = strings.TrimSpace(arl)

	// Basic validation: ARL tokens are typically long strings
	if len(arl) < 100 {
		return fmt.Errorf("ARL token appears to be invalid (too short)")
	}

	// Check if it contains only valid characters (alphanumeric and some special chars)
	for _, char := range arl {
		if !isValidARLChar(char) {
			return fmt.Errorf("ARL token contains invalid characters")
		}
	}

	return nil
}

func isValidARLChar(r rune) bool {
	return (r >= 'a' && r <= 'z') ||
		(r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9') ||
		r == '_' || r == '-' || r == '.' || r == '~'
}

// SaveARLToEnv saves the ARL token to a .env file
func SaveARLToEnv(arl string) error {
	envPath := ".env"
	
	// Read existing .env file if it exists
	content := ""
	if data, err := os.ReadFile(envPath); err == nil {
		content = string(data)
	}

	// Check if DEEZER_ARL already exists
	lines := strings.Split(content, "\n")
	found := false
	for i, line := range lines {
		if strings.HasPrefix(line, "DEEZER_ARL=") {
			lines[i] = fmt.Sprintf("DEEZER_ARL=%s", arl)
			found = true
			break
		}
	}

	// If not found, append it
	if !found {
		if len(lines) > 0 && lines[len(lines)-1] != "" {
			lines = append(lines, "")
		}
		lines = append(lines, fmt.Sprintf("DEEZER_ARL=%s", arl))
	}

	// Write back to file
	newContent := strings.Join(lines, "\n")
	return os.WriteFile(envPath, []byte(newContent), 0644)
}