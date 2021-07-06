package actions

import (
	"fmt"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/libbeat/processors"
)

type dataid struct {
	dataID dataidConfig
}

type dataidConfig struct {
	DataID uint64 `config:"dataid"`
}

// init registers the add_cloud_metadata processor.
func init() {
	processors.RegisterPlugin("set_dataid", configChecked(new, requireFields("dataid"), allowedFields("dataid", "when")))
}

func new(c *common.Config) (processors.Processor, error) {
	processor := &dataid{dataID: dataidConfig{DataID: 0}}
	err := c.Unpack(&processor.dataID)
	if err != nil {
		return nil, fmt.Errorf("fail to unpack the drop_fields configuration: %s", err)
	}

	logp.Debug("bkfilters", "bkfilter config %v", processor.dataID)

	return processor, nil
}

func (cli dataid) Run(event *beat.Event) (*beat.Event, error) {
	data := event.Fields
	if ok, _ := data.HasKey("dataid"); !ok {

		data.Update(common.MapStr{"dataid": cli.dataID.DataID})
	}
	event.Fields = data
	return event, nil
}

func (cli dataid) String() string {

	return "set_dataid"
}
