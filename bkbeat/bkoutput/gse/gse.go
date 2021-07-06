package gse

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/TencentBlueKing/collector-go-sdk/bkbeat/bkmonitoring"
	bkmonitoringReport "github.com/TencentBlueKing/collector-go-sdk/bkbeat/bkmonitoring/report/bkpipe"
	"github.com/TencentBlueKing/collector-go-sdk/bkbeat/gselib"
	bkstorage "github.com/TencentBlueKing/collector-go-sdk/bkbeat/storage"
	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/libbeat/outputs"
	"github.com/elastic/beats/libbeat/publisher"
)

const maxSyncAgentInfoTimeout = 10 // unit: second

var (
	metricGseTaskPublishTotal  string = "gse_publish_total"  //按任务计算发送次数
	metricGseTaskPublishFailed string = "gse_publish_failed" //按任务计算发送失败次数
)

var (
	metricGseAgentInfoFailed = bkmonitoring.NewInt("gse_agent_info_failed") //获取Agent失败次数
	metricGseSendTotal       = bkmonitoring.NewInt("gse_send_total")        //发送给gse client的事件数

	metricGsePublishReceived = bkmonitoring.NewInt("gse_publish_received") //publish：接收事件数
	metricGsePublishDropped  = bkmonitoring.NewInt("gse_publish_dropped")  //publish：丢失的事件数（缺少dataid）
	metricGsePublishTotal    = bkmonitoring.NewInt("gse_publish_total")    //publish：发送事件数
	metricGsePublishFailed   = bkmonitoring.NewInt("gse_publish_failed")   //publish：发送失败数

	metricGseReportReceived  = bkmonitoring.NewInt("gse_report_received")   //report：接收事件数
	metricGseReportSendTotal = bkmonitoring.NewInt("gse_report_send_total") //report：发送事件数
	metricGseReportFailed    = bkmonitoring.NewInt("gse_report_failed")     //report：发送失败数
)

func init() {
	outputs.RegisterType("gse", MakeGSE)
}

// Output : gse output, for libbeat output
type Output struct {
	cli *gselib.GseClient
}

// MakeGSE create a gse client
func MakeGSE(im outputs.IndexManager, beat beat.Info, stats outputs.Observer, cfg *common.Config) (outputs.Group, error) {
	c := defaultConfig
	err := cfg.Unpack(&c)
	if err != nil {
		logp.Err("unpack config error, %v", err)
		return outputs.Fail(err)
	}
	logp.Info("gse config: %+v", c)

	// create gse client
	cli, err := gselib.NewGseClient(cfg)
	if err != nil {
		return outputs.Fail(err)
	}
	output := &Output{
		cli: cli,
	}

	// start gse client
	err = output.cli.Start()
	if err != nil {
		logp.Err("init output failed, %v", err)
		return outputs.Fail(err)
	}
	logp.Info("start gse output")

	// wait to get agent info
	agentInfo, err := output.cli.GetAgentInfo()
	count := maxSyncAgentInfoTimeout
	for {
		if count <= 0 {
			return outputs.Fail(fmt.Errorf("get agent info timeout"))
		}
		if agentInfo.IP != "" {
			break
		}
		count--
		// sleep 1s, then continue to get agent info
		time.Sleep(1 * time.Second)
		agentInfo, err = output.cli.GetAgentInfo()
	}

	// init bkmonitoring
	bkmonitoringReport.InitSender(output, int(agentInfo.Cloudid), agentInfo.IP)

	return outputs.Success(int(c.EventBufferMax), 0, output)
}

// Publish implement output interface
func (c *Output) Publish(batch publisher.Batch) error {
	events := batch.Events()
	var errMsg string
	for i := range events {
		if events[i].Content.Fields == nil {
			metricGsePublishDropped.Add(1)
			continue
		}
		metricGsePublishReceived.Add(1)
		err := c.PublishEvent(&events[i])
		if err != nil {
			errMsg += fmt.Sprintf("Event %d: ", i) + err.Error() + "; "
			metricGsePublishFailed.Add(1)
		} else {
			metricGsePublishTotal.Add(1)
		}
	}

	batch.ACK()

	if errMsg == "" {
		return nil
	}
	return fmt.Errorf(errMsg)
}

// String returns the name of the output client
func (c *Output) String() string {
	return "gse"
}

// PublishEvent implement output interface
// data is event, must contain 'dataid' filed
// data will attach agent info, see publishEventAttachInfo
func (c *Output) PublishEvent(event *publisher.Event) error {
	// get dataid from event
	content := event.Content
	data := content.Fields
	val, err := data.GetValue("dataid")
	if err != nil {
		logp.Err("event lost dataid field, %v", err)
		return fmt.Errorf("event lost dataid")
	}

	dataid := c.getdataid(val)
	if dataid <= 0 {
		return fmt.Errorf("dataid %d <= 0", dataid)
	}

	if content.Meta != nil {
		data.Put("@meta", content.Meta)
	}

	if err := c.publishEventAttachInfo(dataid, data); err != nil {
		return err
	}

	return nil
}

// Close : close gse out put
func (c *Output) Close() error {
	logp.Err("gse output close")
	c.cli.Close()
	return nil
}

// publishEventAttachInfo attach agentinfo and gseindex
// will add bizid, cloudid, ip, gseindex
func (c *Output) publishEventAttachInfo(dataid int32, data common.MapStr) error {
	// 是否兼容原采集器输出
	isStandardFormat := true
	if _, ok := data["_time_"]; ok {
		isStandardFormat = false
	}

	// add gseindex
	var gseIndex uint64
	if ok, _ := data.HasKey("gseindex"); !ok {
		gseIndex = getGseIndex(dataid)
	}

	// add bizid, cloudid, ip
	info, _ := c.cli.GetAgentInfo()
	if len(info.IP) == 0 {
		metricGseAgentInfoFailed.Add(1)
		return fmt.Errorf("agent info is empty")
	}

	if isStandardFormat {
		data["bizid"] = info.Bizid
		data["cloudid"] = info.Cloudid
		data["ip"] = info.IP
		data["hostname"] = info.Hostname
		data["gseindex"] = gseIndex
	} else {
		data["_bizid_"] = info.Bizid
		data["_cloudid_"] = info.Cloudid
		data["_server_"] = info.IP
		data["_gseindex_"] = gseIndex
		data.Delete("dataid")
		data.Delete("time")
	}

	return c.reportCommonData(dataid, data)
}

func getGseIndex(dataid int32) uint64 {
	index := uint64(0)
	gseIndexKey := fmt.Sprintf("gseindex_%s", String(dataid))
	if indexStr, err := bkstorage.Get(gseIndexKey); nil == err {
		if index, err = strconv.ParseUint(indexStr, 10, 64); nil != err {
			logp.Err("fail to get gseindex %v", err)
			index = 0
		}
	}
	index++
	bkstorage.Set(gseIndexKey, fmt.Sprintf("%v", index), 0)
	return index
}

// Report implement interface for bkmonitor
func (c *Output) Report(dataid int32, data common.MapStr) error {
	if dataid <= 0 {
		return fmt.Errorf("dataid %d <= 0", dataid)
	}
	metricGseReportReceived.Add(1)
	err := c.reportCommonData(dataid, data)
	if err != nil {
		metricGseReportFailed.Add(1)
		return err
	}
	metricGseReportSendTotal.Add(1)
	return nil
}

// ReportRaw implement interface for monitor
// send op raw data, without attach anything
func (c *Output) ReportRaw(dataid int32, data interface{}) error {
	if dataid <= 0 {
		return fmt.Errorf("dataid %d <= 0", dataid)
	}

	buf, err := json.Marshal(data)
	if err != nil {
		logp.Err("convert to json failed: %v", err)
		return err
	}

	logp.Debug("gse", "report data to %d", dataid)
	// report op data

	msg := gselib.NewGseOpMsg(buf, dataid, 0, 0, 0)
	logp.Debug("gse", "report data : %s", string(buf))
	// TODO compatible op data bug fixed after agent D48
	// send every op data with new connection
	c.cli.SendWithNewConnection(msg)
	//c.cli.Send(msg)

	return nil
}

// reportCommonData send common data
func (c *Output) reportCommonData(dataid int32, data common.MapStr) error {
	// change data to json format
	buf, err := json.Marshal(data)
	if err != nil {
		bkmonitoring.NewIntWithDataID(int(dataid), metricGseTaskPublishFailed).Add(1)
		logp.Err("json marshal failed, %v", err)
		return err
	}

	// new dynamic msg
	msg := gselib.NewGseDynamicMsg(buf, dataid, 0, 0)

	// send data
	c.cli.Send(msg)

	// 发包计数
	metricGseSendTotal.Add(1)
	bkmonitoring.NewIntWithDataID(int(dataid), metricGseTaskPublishTotal).Add(1)

	return nil
}

func (c *Output) getdataid(dataID interface{}) int32 {
	switch dataID.(type) {
	case int, int8, int16, int32, int64:
		return int32(reflect.ValueOf(dataID).Int())
	case uint, uint8, uint16, uint32, uint64:
		return int32(reflect.ValueOf(dataID).Uint())
	case string:
		dataid, err := strconv.ParseInt(dataID.(string), 10, 32)
		if err != nil {
			logp.Err("can not get dataid, %s", dataID.(string))
			return -1
		}
		return int32(dataid)
	default:
		logp.Err("unexpected type %T for the dataid ", dataID)
		return 0
	}
}

func String(n int32) string {
	buf := [11]byte{}
	pos := len(buf)
	i := int64(n)
	signed := i < 0
	if signed {
		i = -i
	}
	for {
		pos--
		buf[pos], i = '0'+byte(i%10), i/10
		if i == 0 {
			if signed {
				pos--
				buf[pos] = '-'
			}
			return string(buf[pos:])
		}
	}
}
