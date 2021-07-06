package beat

import (
	"fmt"
	"sync"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
)

// BeaterState is beater's state
type BeaterState int

const (
	BeaterBeforeOpening BeaterState = iota
	BeaterFailToOpen
	BeaterRunning
	BeaterStoped
)

// BKBeat
type BKBeat struct {
	Beat        *beat.Beat
	Client      *beat.Client
	LocalConfig *Config
	BeaterState BeaterState
	Finished    bool
	Done        chan struct{}
}

var commonBKBeat = BKBeat{
	Finished:    false,
	BeaterState: BeaterBeforeOpening,
	Done:        make(chan struct{}),
}

var publishConfig PublishConfig
var wg sync.WaitGroup
var errorMessageChan chan (error)

func creator(b *beat.Beat, localConfig *common.Config) (beat.Beater, error) {
	commonBKBeat.Beat = b
	commonBKBeat.LocalConfig = localConfig
	if nil == b || nil == localConfig {
		return nil, fmt.Errorf("%s failed to initialize", b.Info.Beat)
	}
	return &commonBKBeat, nil
}

// Run
func (bkb *BKBeat) Run(b *beat.Beat) error {
	bkb.Beat = b
	var err error
	if publishConfig.ACKEvents != nil {
		err = commonBKBeat.Beat.Publisher.SetACKHandler(beat.PipelineACKHandler{
			ACKEvents: publishConfig.ACKEvents,
		})
		if err != nil {
			bkb.Client = nil
			return err
		}
	}
	client, err := commonBKBeat.Beat.Publisher.ConnectWith(publishConfig)
	if nil != err {
		bkb.Client = nil
		return err
	}
	bkb.Client = &client
	wg.Done()
	close(errorMessageChan)
	select {
	case <-bkb.Done:
		bkb.Finished = true
		return nil
	}
}

// Stop
func (bkb *BKBeat) Stop() {
	if BeaterRunning != commonBKBeat.BeaterState {
		return
	}
	if nil == bkb.Client {
		return
	}
	(*bkb.Client).Close()
	bkb.Client = nil
	close(bkb.Done)
	freeResource()
}

// Reload
func (bkb *BKBeat) Reload(localConfig *common.Config) {
	commonBKBeat.LocalConfig = localConfig
	ReloadChan <- true
}

func beatNotRunning() bool {
	return commonBKBeat.BeaterState != BeaterRunning
}
