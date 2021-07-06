package storage

import (
	"errors"
	"time"
)

var _storage Storage
var ErrNotFound = errors.New("key not found")

// StorageConfig
type StorageConfig struct {
}

// Storage : key value storage
type Storage interface {
	Set(key, value string, expire time.Duration) error
	Get(key string) (string, error)
	Del(key string) error
	Close() error
	Destory() error // clean files
}

// Init : init storage
// path is a filepath
func Init(path string, config *StorageConfig) error {
	var err error
	_storage, err = NewLocalStorage(path)
	return err
}

// Get get value. will return error=ErrNotFound if key not exist
func Get(key string) (string, error) {
	return _storage.Get(key)
}

// Set kv to storage, expire not used now
func Set(key, value string, expire time.Duration) error {
	if len(key) == 0 || len(value) == 0 {
		return nil
	}
	return _storage.Set(key, value, expire)
}

// Del delete key, will do nothing if key not exist
func Del(key string) error {
	return _storage.Del(key)
}

func Close() {
	_storage.Close()
}

// Destory remove files
func Destory() error {
	return _storage.Destory()
}
