package cache

import (
	"os"
	"testing"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/stretchr/testify/assert"
)

func TestSetAndGetCache(t *testing.T) {
	key := "test-key"
	value := []byte("test-value")

	// Test setting a cache value
	err := SetCache(key, value)
	assert.NoError(t, err, "expected no error when setting cache value")

	// Test getting a cache value
	cachedValue, err := GetCache(key)
	assert.NoError(t, err, "expected no error when getting cache value")
	assert.Equal(t, value, cachedValue, "expected cached value to match the set value")
}

func TestFlushCache(t *testing.T) {
	// Test flushing the cache
	err := FlushCache()
	assert.NoError(t, err, "expected no error when flushing cache")
}

func TestCloseCache(t *testing.T) {
	// Test closing the cache
	err := CloseCache()
	assert.NoError(t, err, "expected no error when closing cache")

	// Reinitialize cache after closing for other tests
	initCache()
}

func initCache() {
	cacheDir := "./test_cache"
	opts := badger.DefaultOptions(cacheDir).
		WithLogger(nil).
		WithBlockCacheSize(64 * 1024 * 1024).
		WithIndexCacheSize(32 * 1024 * 1024).
		WithMemTableSize(32 * 1024 * 1024)

	var err error
	cache, err = badger.Open(opts)
	if err != nil {
		panic("Failed to reinitialize cache: " + err.Error())
	}

	// Cleanup after tests
	go func() {
		time.Sleep(1 * time.Second) // Wait a moment before cleanup
		_ = CloseCache()
		_ = os.RemoveAll(cacheDir)
	}()
}

func TestCacheExpiry(t *testing.T) {
	key := "expire-key"
	value := []byte("expire-value")

	// Set cache with a short TTL for testing expiration
	originalTTL := ttl
	ttl = 1 * time.Second
	err := SetCache(key, value)
	assert.NoError(t, err, "expected no error when setting cache value with TTL")

	// Sleep to let the cache entry expire
	time.Sleep(2 * time.Second)

	// Attempt to get the expired cache value
	cachedValue, err := GetCache(key)
	assert.Error(t, err, "expected error when getting expired cache value")
	assert.Nil(t, cachedValue, "expected no value for expired cache key")

	// Reset TTL to original value after test
	ttl = originalTTL
}
