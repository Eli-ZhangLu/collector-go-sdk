package bkpipe

import (
	"fmt"
	"testing"

	"github.com/TencentBlueKing/collector-go-sdk/v2/bkbeat/bkmonitoring"
	"github.com/TencentBlueKing/collector-go-sdk/v2/bkbeat/logp"

	"github.com/elastic/beats/libbeat/common"
	libbeatlogp "github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/libbeat/monitoring"
	"github.com/stretchr/testify/assert"
)

func init() {
	logp.SetLogger(libbeatlogp.L())
}

var (
	beatName     string = "bkbeat"
	beatVersion  string = "v1"
	dataID       int32  = 999999
	taskDataID   int32  = 888888
	metricDataID int    = 666666
	bkCloudId    int    = 0
	ip           string = "1.1.1.1"
)

type mockSender struct {
	Data map[int32]common.MapStr
}

func (client *mockSender) Report(dataid int32, data common.MapStr) error {
	client.Data[dataid] = data
	return nil
}

func TestMetrics(t *testing.T) {
	//采集器公共指标
	metric1 := bkmonitoring.NewInt("test.v", monitoring.Gauge)
	metric1.Add(1)

	//采集器任务指标
	metric2 := bkmonitoring.NewIntWithDataID(metricDataID, "v")
	metric2.Add(1)
	last := makeSnapshot(monitoring.Default)

	metric1.Add(1)
	metric2.Add(1)
	cur := makeSnapshot(monitoring.Default)

	r := &reporter{
		dataID:      dataID,
		taskDataID:  taskDataID,
		registry:    monitoring.Default,
		beatName:    beatName,
		beatVersion: beatVersion,
	}

	output := &mockSender{
		Data: map[int32]common.MapStr{
			dataID:     common.MapStr{},
			taskDataID: common.MapStr{},
		},
	}
	InitSender(output, bkCloudId, ip)
	delta := r.makeDeltaSnapshot(last, cur)

	var delta1 int64 = 1
	var delta2 int64 = 2
	assert.Equal(t, delta.Ints["bkbeat.test.v"], delta2)
	taskKey := fmt.Sprintf("bkbeat_tasks.%d.v", metricDataID)
	assert.Equal(t, delta.Ints[taskKey], delta1)
	r.sendMetrics(delta)

	assertMetric := map[int32]common.MapStr{
		dataID: common.MapStr{
			"metrics": common.MapStr{
				"bkbeat_test_v": delta2,
			},
			"dimension": common.MapStr{
				"type":         beatName,
				"version":      beatVersion,
				"task_data_id": 0,
			},
		},
		taskDataID: common.MapStr{
			"metrics": common.MapStr{
				"v": delta1,
			},
			"dimension": common.MapStr{
				"type":         beatName,
				"version":      beatVersion,
				"task_data_id": metricDataID,
			},
		},
	}

	for k, v := range output.Data {
		item := v["data"].([]common.MapStr)
		assert.Equal(t, item[0], assertMetric[k])
	}
}
