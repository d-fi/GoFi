package api

import (
	"log"
	"time"

	"github.com/dgraph-io/badger/v4"
)

var (
	cache *badger.DB
	ttl   = 60 * time.Minute
)

func init() {
	var err error
	opts := badger.DefaultOptions("").
		WithInMemory(true).
		WithLogger(nil).
		WithBlockCacheSize(16 * 1024 * 1024). // Set block cache size to 16 MB
		WithIndexCacheSize(8 * 1024 * 1024).  // Set index cache size to 8 MB
		WithMemTableSize(8 * 1024 * 1024)     // Set MemTable size to 8 MB

	cache, err = badger.Open(opts)
	if err != nil {
		log.Fatalf("Failed to initialize cache: %v", err)
	}

}

func setCache(key string, value []byte) error {
	return cache.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry([]byte(key), value).WithTTL(ttl)
		return txn.SetEntry(e)
	})
}

func getCache(key string) ([]byte, error) {
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
