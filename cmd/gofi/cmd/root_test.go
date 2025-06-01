package cmd

import (
	"os"
	"testing"
)

func TestGetEnvOrDefault(t *testing.T) {
	tests := []struct {
		name         string
		envKey       string
		envValue     string
		defaultValue string
		expected     string
	}{
		{
			name:         "Environment variable set",
			envKey:       "TEST_ENV_VAR",
			envValue:     "test_value",
			defaultValue: "default",
			expected:     "test_value",
		},
		{
			name:         "Environment variable not set",
			envKey:       "TEST_ENV_VAR_NOT_SET",
			envValue:     "",
			defaultValue: "default",
			expected:     "default",
		},
		{
			name:         "Empty environment variable",
			envKey:       "TEST_ENV_EMPTY",
			envValue:     "",
			defaultValue: "default",
			expected:     "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set or unset environment variable
			if tt.envValue != "" {
				os.Setenv(tt.envKey, tt.envValue)
				defer os.Unsetenv(tt.envKey)
			} else {
				os.Unsetenv(tt.envKey)
			}

			// Test the function
			result := getEnvOrDefault(tt.envKey, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getEnvOrDefault() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetEnvIntOrDefault(t *testing.T) {
	tests := []struct {
		name         string
		envKey       string
		envValue     string
		defaultValue int
		expected     int
		expectWarn   bool
	}{
		{
			name:         "Valid integer",
			envKey:       "TEST_INT_VAR",
			envValue:     "42",
			defaultValue: 10,
			expected:     42,
			expectWarn:   false,
		},
		{
			name:         "Invalid integer",
			envKey:       "TEST_INT_INVALID",
			envValue:     "not_a_number",
			defaultValue: 10,
			expected:     10,
			expectWarn:   true,
		},
		{
			name:         "Environment variable not set",
			envKey:       "TEST_INT_NOT_SET",
			envValue:     "",
			defaultValue: 10,
			expected:     10,
			expectWarn:   false,
		},
		{
			name:         "Zero value",
			envKey:       "TEST_INT_ZERO",
			envValue:     "0",
			defaultValue: 10,
			expected:     0,
			expectWarn:   false,
		},
		{
			name:         "Negative value",
			envKey:       "TEST_INT_NEGATIVE",
			envValue:     "-5",
			defaultValue: 10,
			expected:     -5,
			expectWarn:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set or unset environment variable
			if tt.envValue != "" {
				os.Setenv(tt.envKey, tt.envValue)
				defer os.Unsetenv(tt.envKey)
			} else {
				os.Unsetenv(tt.envKey)
			}

			// Test the function
			result := getEnvIntOrDefault(tt.envKey, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getEnvIntOrDefault() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestEnvironmentVariableIntegration(t *testing.T) {
	// Save original values
	origOutput := os.Getenv("GOFI_OUTPUT_DIR")
	origQuality := os.Getenv("GOFI_QUALITY")
	origLogLevel := os.Getenv("GOFI_LOG_LEVEL")

	// Restore original values after test
	defer func() {
		if origOutput != "" {
			os.Setenv("GOFI_OUTPUT_DIR", origOutput)
		} else {
			os.Unsetenv("GOFI_OUTPUT_DIR")
		}
		if origQuality != "" {
			os.Setenv("GOFI_QUALITY", origQuality)
		} else {
			os.Unsetenv("GOFI_QUALITY")
		}
		if origLogLevel != "" {
			os.Setenv("GOFI_LOG_LEVEL", origLogLevel)
		} else {
			os.Unsetenv("GOFI_LOG_LEVEL")
		}
	}()

	tests := []struct {
		name            string
		outputDir       string
		quality         string
		logLevel        string
		expectedOutput  string
		expectedQuality int
		expectedLog     string
	}{
		{
			name:            "All environment variables set",
			outputDir:       "/tmp/test-music",
			quality:         "9",
			logLevel:        "debug",
			expectedOutput:  "/tmp/test-music",
			expectedQuality: 9,
			expectedLog:     "debug",
		},
		{
			name:            "Invalid quality - should use default",
			outputDir:       "/tmp/test",
			quality:         "7",
			logLevel:        "info",
			expectedOutput:  "/tmp/test",
			expectedQuality: 3,
			expectedLog:     "info",
		},
		{
			name:            "No environment variables - use defaults",
			outputDir:       "",
			quality:         "",
			logLevel:        "",
			expectedOutput:  "./downloads",
			expectedQuality: 3,
			expectedLog:     "info",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			if tt.outputDir != "" {
				os.Setenv("GOFI_OUTPUT_DIR", tt.outputDir)
			} else {
				os.Unsetenv("GOFI_OUTPUT_DIR")
			}
			if tt.quality != "" {
				os.Setenv("GOFI_QUALITY", tt.quality)
			} else {
				os.Unsetenv("GOFI_QUALITY")
			}
			if tt.logLevel != "" {
				os.Setenv("GOFI_LOG_LEVEL", tt.logLevel)
			} else {
				os.Unsetenv("GOFI_LOG_LEVEL")
			}

			// Test getEnvOrDefault for output dir
			output := getEnvOrDefault("GOFI_OUTPUT_DIR", "./downloads")
			if output != tt.expectedOutput {
				t.Errorf("Expected output dir %s, got %s", tt.expectedOutput, output)
			}

			// Test getEnvIntOrDefault for quality
			quality := getEnvIntOrDefault("GOFI_QUALITY", 3)
			// Apply validation as in init()
			if quality != 1 && quality != 3 && quality != 9 {
				quality = 3
			}
			if quality != tt.expectedQuality {
				t.Errorf("Expected quality %d, got %d", tt.expectedQuality, quality)
			}

			// Test getEnvOrDefault for log level
			logLevel := getEnvOrDefault("GOFI_LOG_LEVEL", "info")
			if logLevel != tt.expectedLog {
				t.Errorf("Expected log level %s, got %s", tt.expectedLog, logLevel)
			}
		})
	}
}