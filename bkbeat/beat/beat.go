package beat

import (
	"flag"
	"time"

	// important plugins
	_ "github.com/TencentBlueKing/collector-go-sdk/bkbeat/bkmonitoring/report/bkpipe"
	_ "github.com/TencentBlueKing/collector-go-sdk/bkbeat/bkoutput/bkpipe"
	_ "github.com/TencentBlueKing/collector-go-sdk/bkbeat/bkoutput/gse"
	_ "github.com/TencentBlueKing/collector-go-sdk/bkbeat/bkprocessor/actions"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
)

type MapStr = common.MapStr

type Event = beat.Event

type PublishConfig = beat.ClientConfig

type ProcessingConfig = beat.ProcessingConfig

type PublishMode = beat.PublishMode

type ClientEventer = beat.ClientEventer

type EventMetadata = common.EventMetadata

type MapStrPorinter = common.MapStrPointer

type ProcessorList = beat.ProcessorList

type Processor = beat.Processor

const DefaultGuarantees = beat.DefaultGuarantees
const GuaranteedSend = beat.GuaranteedSend
const DropIfFull = beat.DropIfFull

// ReloadChan indicates developers to reload config when new config is ready
var ReloadChan chan bool
var Done chan bool

var (
	reloadFlag = flag.Bool("reload", false, "Reload the program")
)

func bkEventToEvent(data MapStr) beat.Event {
	ev := beat.Event{
		Fields:    data,
		Timestamp: time.Now(),
	}
	return ev
}

func formatEvent(event beat.Event) beat.Event {
	if event.Fields == nil {
		return event
	}

	if _, ok := event.Fields["time"]; !ok {
		event.Timestamp = time.Now()
		event.Fields.Put("time", event.Timestamp.Format("2006-01-02 15:04:05"))
	}

	return event
}
