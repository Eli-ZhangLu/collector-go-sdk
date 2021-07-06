package pidfile

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	bkcommon "github.com/TencentBlueKing/collector-go-sdk/bkbeat/common"
	"github.com/nightlyone/lockfile"
)

var lock lockfile.Lockfile

// GetPid get pid from pidfile
func GetPid(pidFilePath string) (int, error) {
	// read file
	buf, err := ioutil.ReadFile(pidFilePath)
	if err != nil {
		return -1, err
	}
	pid := bkcommon.ScanPidLine(buf)
	if pid <= 0 {
		return -1, fmt.Errorf("can not get pid!")
	}
	return int(pid), err
}

// TryLock try to create lockfile
func TryLock(pidFilePath string) error {
	// ensure pid path exist
	dir := filepath.Dir(pidFilePath)
	err := os.MkdirAll(dir, 0775)
	if err != nil {
		fmt.Printf("Cannot create pid directory. reason: %v", err)
		return err
	}

	lock, err = lockfile.New(pidFilePath)
	if err != nil {
		fmt.Printf("Cannot init lock. reason: %v", err)
		return err
	}
	err = lock.TryLock()
	// Error handling is essential, as we only try to get the lock.
	if err != nil {
		fmt.Printf("Cannot lock %q, reason: %v", lock, err)
		return err
	}
	return nil
}

func UnLock() {
	lock.Unlock()
}
