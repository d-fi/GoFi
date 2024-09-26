package request

import (
	"testing"
)

func TestSetCache(t *testing.T) {
	key := "key1"
	value := []byte("value1")

	err := setCache(key, value)
	if err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}

	retrievedValue, err := getCache(key)
	if err != nil {
		t.Fatalf("Failed to get cache: %v", err)
	}

	if string(retrievedValue) != string(value) {
		t.Fatalf("Expected value %s, got %s", value, retrievedValue)
	}
}

func TestGetCache(t *testing.T) {
	key := "nonExistentKey"

	_, err := getCache(key)
	if err == nil {
		t.Fatalf("Expected an error for non-existent key, got nil")
	}
}
