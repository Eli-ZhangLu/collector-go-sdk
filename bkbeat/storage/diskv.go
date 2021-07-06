// +build aix

package storage

import (
	"os"
	"path/filepath"
	"time"

	"github.com/peterbourgon/diskv"
)

type LocalStorage struct {
	path  string
	store *diskv.Diskv
}

// new dikv storage
func NewLocalStorage(path string) (*LocalStorage, error) {
	flatTransform := func(s string) []string { return []string{} }
	kvPath := filepath.Join(filepath.Dir(path), "dikv")
	store := diskv.New(diskv.Options{
		BasePath:     kvPath,
		Transform:    flatTransform,
		CacheSizeMax: 1024 * 1024,
	})
	return &LocalStorage{store: store, path: kvPath}, nil
}

// Close : close db
func (cli *LocalStorage) Close() error {
	return nil
}

// Set : set value
func (cli *LocalStorage) Set(key, value string, expire time.Duration) error {
	return cli.store.WriteString(key, value)
}

// Get : get value
// if not found, return ErrNotFound
func (cli *LocalStorage) Get(key string) (string, error) {
	buf, err := cli.store.Read(key)
	if len(buf) == 0 {
		err = ErrNotFound
	}
	return string(buf), err
}

// Del : delete key
func (cli *LocalStorage) Del(key string) error {
	return cli.store.Erase(key)
}

// Destory remove db files
func (cli *LocalStorage) Destory() error {
	cli.Close()
	return os.RemoveAll(cli.path)
}
