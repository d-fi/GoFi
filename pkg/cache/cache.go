package cache

import (
	"log"
	"time"

	"github.com/d-fi/GoFi/pkg/utils"
	"github.com/dgraph-io/badger/v4"
)

var (
	cache *badger.DB
	ttl   = 60 * time.Minute
)

func init() {
	cacheDir := utils.GetSystemCacheDir("GoFi") + "/badger"
	opts := badger.DefaultOptions(cacheDir).
		WithLogger(nil).
		WithBlockCacheSize(64 * 1024 * 1024). // 64 MB block cache
		WithIndexCacheSize(32 * 1024 * 1024). // 32 MB index cache
		WithMemTableSize(32 * 1024 * 1024)    // 32 MB MemTable

	var err error
	cache, err = badger.Open(opts)
	if err != nil {
		log.Fatalf("Failed to initialize cache: %v", err)
	}
}

// SetCache sets a value in the cache with the specified TTL.
func SetCache(key string, value []byte) error {
	return cache.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry([]byte(key), value).WithTTL(ttl)
		return txn.SetEntry(e)
	})
}

// GetCache retrieves a value from the cache.
func GetCache(key string) ([]byte, error) {
	var valCopy []byte
	err := cache.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			valCopy = append([]byte{}, val...)
			return nil
		})
	})
	return valCopy, err
}

// FlushCache flushes the pending writes to the disk.
func FlushCache() error {
	return cache.Flatten(0)
}

// CloseCache closes the Badger DB instance gracefully.
func CloseCache() error {
	return cache.Close()
}
