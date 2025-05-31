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
		return cleanARLToken(arl), nil
	}

	// Try to get from browser cookies
	arl, err := GetARLFromAnyBrowser()
	if err == nil && arl != "" {
		return cleanARLToken(arl), nil
	}

	// Return error if no ARL found
	return "", fmt.Errorf("ARL token not found in environment or browser cookies: %w", err)
}

// cleanARLToken removes any control characters from the ARL token
func cleanARLToken(arl string) string {
	// Clean the ARL token - remove any control characters
	cleanARL := ""
	for _, r := range arl {
		if r >= 32 && r <= 126 {
			cleanARL += string(r)
		}
	}
	return strings.TrimSpace(cleanARL)
}

// ValidateARLToken performs basic validation on an ARL token
func ValidateARLToken(arl string) error {
	// Remove any whitespace
	arl = strings.TrimSpace(arl)

	// Basic validation: ARL tokens are typically long strings
	if len(arl) < 100 {
		return fmt.Errorf("ARL token appears to be invalid (too short)")
	}

	// More lenient validation - just check if it has mostly valid characters
	validChars := 0
	for _, char := range arl {
		if isValidARLChar(char) {
			validChars++
		}
	}
	
	// If at least 90% of characters are valid, accept it
	if float64(validChars) / float64(len(arl)) < 0.9 {
		return fmt.Errorf("ARL token contains too many invalid characters")
	}

	return nil
}

func isValidARLChar(r rune) bool {
	// ARL tokens are base64-like strings that can contain:
	// - Letters (a-z, A-Z)
	// - Numbers (0-9)
	// - Special characters used in base64 and URL encoding
	return (r >= 'a' && r <= 'z') ||
		(r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9') ||
		r == '_' || r == '-' || r == '.' || r == '~' ||
		r == '+' || r == '/' || r == '=' || r == '%'
}

// SaveARLToEnv saves the ARL token to a .env file
func SaveARLToEnv(arl string) error {
	envPath := ".env"
	
	// Clean the ARL token - remove any control characters
	cleanARL := ""
	for _, r := range arl {
		if r >= 32 && r <= 126 {
			cleanARL += string(r)
		}
	}
	
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
			lines[i] = fmt.Sprintf("DEEZER_ARL=%s", cleanARL)
			found = true
			break
		}
	}

	// If not found, append it
	if !found {
		if len(lines) > 0 && lines[len(lines)-1] != "" {
			lines = append(lines, "")
		}
		lines = append(lines, fmt.Sprintf("DEEZER_ARL=%s", cleanARL))
	}

	// Write back to file
	newContent := strings.Join(lines, "\n")
	return os.WriteFile(envPath, []byte(newContent), 0644)
}