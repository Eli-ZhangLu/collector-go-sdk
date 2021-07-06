package bkpipe

import (
	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/outputs"

	"github.com/TencentBlueKing/collector-go-sdk/bkbeat/bkoutput/gse"
)

func init() {
	outputs.RegisterType("bkpipe", MakeBKPipe)
}

// MakeBKPipe create gse output
// compatible with old configurations
func MakeBKPipe(im outputs.IndexManager, beat beat.Info, stats outputs.Observer, cfg *common.Config) (outputs.Group, error) {
	return gse.MakeGSE(im, beat, stats, cfg)
}
