package beat

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	bkcommon "github.com/TencentBlueKing/collector-go-sdk/bkbeat/common"
	"github.com/TencentBlueKing/collector-go-sdk/bkbeat/logp"
	"github.com/TencentBlueKing/collector-go-sdk/bkbeat/pidfile"
	bkreloader "github.com/TencentBlueKing/collector-go-sdk/bkbeat/reloader"
	bkstorage "github.com/TencentBlueKing/collector-go-sdk/bkbeat/storage"

	"github.com/elastic/beats/libbeat/cfgfile"
	"github.com/elastic/beats/libbeat/cmd/instance"
	"github.com/elastic/beats/libbeat/common"
	libbeatlogp "github.com/elastic/beats/libbeat/logp"
)

type pathConfig struct {
	PidFilePath string `config:"pid"`
	DataPath    string `config:"data"`
}

const pathConfigField string = "path"

var defaultPathConfig = pathConfig{
	PidFilePath: "",
	DataPath:    "./data",
}

var reloader *bkreloader.Reloader
var mutex sync.Mutex
var rawconfig *Config
var beatSettings instance.Settings

func baseInit(beatName string, version string) (*Config, error) {
	rawconfig = nil
	ReloadChan = make(chan bool)
	Done = make(chan bool)
	commonBKBeat = BKBeat{
		Finished:    false,
		BeaterState: BeaterBeforeOpening,
		Done:        make(chan struct{}),
	}
	cfgfile.ChangeDefaultCfgfileFlag(beatName)

	flag.Parse()

	// Print version
	versionFlag := flag.Lookup("v")
	if versionFlag.Value.String() == "true" {
		fmt.Printf("%s\n", version)
		os.Exit(0)
	}

	// Get Pid file path
	var err error
	rawconfig, err = cfgfile.Load("", nil)
	if err != nil {
		commonBKBeat.BeaterState = BeaterFailToOpen
		return nil, err
	}
	var pathCfg *common.Config
	pathConfig := pathConfig(defaultPathConfig)
	if rawconfig.HasField(pathConfigField) {
		pathCfg, err = rawconfig.Child(pathConfigField, -1)
		if err != nil {
			commonBKBeat.BeaterState = BeaterFailToOpen
			return nil, err
		}
		err = pathCfg.Unpack(&pathConfig)
		if err != nil {
			commonBKBeat.BeaterState = BeaterFailToOpen
			return nil, err
		}
	} else {
		pathConfig.PidFilePath = ""
	}
	pidFilePath, err := bkcommon.MakePifFilePath(beatName, pathConfig.PidFilePath)
	if err != nil {
		commonBKBeat.BeaterState = BeaterFailToOpen
		return nil, err
	}

	// Reload event
	if *reloadFlag {
		err = bkreloader.ReloadEvent(beatName, pidFilePath)
		if err != nil {
			fmt.Println(err.Error())
		}
		os.Exit(0)
	}

	// Lock pid file
	err = pidfile.TryLock(pidFilePath)
	if err != nil {
		commonBKBeat.BeaterState = BeaterFailToOpen
		return nil, err
	}

	// Init bkstorage
	// config is nil now
	dbFilePath := filepath.Join(pathConfig.DataPath, beatName+".bkpipe.db")
	err = bkstorage.Init(dbFilePath, nil)
	if err != nil {
		commonBKBeat.BeaterState = BeaterFailToOpen
		return nil, fmt.Errorf("initializing storage %s error: %v", dbFilePath, err)
	}

	errorMessageChan = make(chan error)
	wg.Add(1)
	// Init libbeat
	go func() {
		beatSettings.Name = beatName
		beatSettings.Version = version
		err := instance.Run(beatSettings, creator)
		if err != nil {
			commonBKBeat.BeaterState = BeaterFailToOpen
			freeResource()
			wg.Done()
			errorMessageChan <- err
			return
		}
		close(Done)
		return
	}()
	wg.Wait()
	err, ok := <-errorMessageChan
	if ok {
		err = fmt.Errorf("failed to initialize libbeat: %s", err.Error())
		fmt.Println(err.Error())
		return nil, err
	}

	logp.SetLogger(libbeatlogp.L())

	// Init reloader
	reloader = bkreloader.NewReloader(beatName, &commonBKBeat)
	if err := reloader.Run(pidFilePath); err != nil {
		logp.L.Errorf(err.Error())
		commonBKBeat.BeaterState = BeaterFailToOpen
		commonBKBeat.Stop()
		return nil, err
	}

	commonBKBeat.BeaterState = BeaterRunning

	return commonBKBeat.LocalConfig, nil
}

// Init initializes user's beat
func Init(beatName string, version string) (*Config, error) {
	mutex.Lock()
	if !beatNotRunning() {
		mutex.Unlock()
		return nil, fmt.Errorf("%s has already been created", beatName)
	}
	config, err := baseInit(beatName, version)
	mutex.Unlock()
	return config, err
}

// InitWithPublishConfig initializes user's beat with user specified publish config
func InitWithPublishConfig(beatName string, version string, pubConfig PublishConfig, settings instance.Settings) (*Config, error) {
	mutex.Lock()
	if !beatNotRunning() {
		mutex.Unlock()
		return nil, fmt.Errorf("%s has already been created", beatName)
	}
	publishConfig = pubConfig
	beatSettings = settings
	config, err := baseInit(beatName, version)
	publishConfig = PublishConfig{}

	mutex.Unlock()
	return config, err
}

func freeResource() {
	pidfile.UnLock()
	bkstorage.Close()
}

// GetConfig fetch the config instance of user's beat
func GetConfig() *Config {
	if beatNotRunning() {
		return nil
	}
	return commonBKBeat.LocalConfig
}

func GetRawConfig() *Config {
	return rawconfig
}

// Send sends a mapstr event
func Send(event MapStr) bool {
	if beatNotRunning() {
		return false
	}
	if nil == commonBKBeat.Client {
		return false
	}
	(*commonBKBeat.Client).Publish(bkEventToEvent(event))
	return true
}

// SendEvent sends a Event type event
func SendEvent(event Event) bool {
	if beatNotRunning() {
		return false
	}
	if nil == commonBKBeat.Client {
		return false
	}
	(*commonBKBeat.Client).Publish(formatEvent(event))
	return true
}

// Stop stops the beat
func Stop() error {
	mutex.Lock()
	if beatNotRunning() {
		mutex.Unlock()
		return errors.New("no beat running")
	}
	commonBKBeat.Stop()
	reloader.Stop()
	freeResource()
	logp.SetLogger(nil)
	commonBKBeat.BeaterState = BeaterStoped
	mutex.Unlock()
	return nil
}
