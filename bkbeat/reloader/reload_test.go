package reloader

import (
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/elastic/beats/libbeat/common"
)

type MockProc struct {
}

var isReload = false

func (*MockProc) Reload(_ *common.Config) {
	isReload = true
}

func Test_reload(t *testing.T) {
	p := &MockProc{}
	name := "bkdata_test"

	// create config file
	cfgFile := "beat.yml"
	os.Remove(cfgFile)
	f, err := os.Create(cfgFile)
	if err != nil {
		t.Fatal(err)
	}
	f.Write([]byte(name + ":"))
	f.Close()
	defer os.Remove(cfgFile)

	// write pid file
	pidFilePath := "/tmp/bkdata_test" // windows will ignore the path
	os.Remove(pidFilePath)
	f, err = os.Create(pidFilePath)
	if err != nil {
		t.Fatal(err)
	}
	f.Write([]byte(strconv.Itoa(os.Getpid())))
	f.Close()

	reloader := NewReloader(name, p)
	if err := reloader.Run(pidFilePath); err != nil {
		t.Fatal(err)
	}
	defer reloader.Stop()
	time.Sleep(1 * time.Second)

	if err := ReloadEvent(name, pidFilePath); err != nil {
		t.Fatal(err)
	}

	time.Sleep(1 * time.Second)
	if !isReload {
		t.Fatal("reload failed")
	}
}
