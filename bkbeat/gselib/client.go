package gselib

import (
	"encoding/binary"
	"encoding/json"
	"io"
	"net"
	"os"
	"reflect"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"

	"github.com/TencentBlueKing/collector-go-sdk/v2/bkbeat/bkmonitoring"
)

// Config GseClient config
type Config struct {
	ReconnectTimes uint          `config:"reconnecttimes"`
	RetryTimes     uint          `config:"retrytimes"`
	RetryInterval  time.Duration `config:"retryinterval"`
	MsgQueueSize   uint          `config:"mqsize"`
	WriteTimeout   time.Duration `config:"writetimeout"`
	Endpoint       string        `config:"endpoint"`
	Nonblock       bool          `config:"nonblock"` // TODO not used now
}

const (
	EINVAL int = iota
	ErrNetClosing
	ErrIOTimeout
	EPIPE
)

var defaultConfig = Config{
	MsgQueueSize:   1,
	WriteTimeout:   5 * time.Second,
	Nonblock:       false,
	RetryTimes:     3,
	RetryInterval:  1 * time.Second,
	ReconnectTimes: 3,
}

var (
	metricGseClientConnected     = bkmonitoring.NewInt("gse_client_connected")      //连接次数
	metricGseClientConnectRetry  = bkmonitoring.NewInt("gse_client_connect_retry")  //连接重试次数
	metricGseClientConnectFailed = bkmonitoring.NewInt("gse_client_connect_failed") //连接失败次数

	metricGseClientReceived    = bkmonitoring.NewInt("gse_client_received")     //接收的请求数
	metricGseClientClose       = bkmonitoring.NewInt("gse_client_close")        //客户端连接断开计数
	metricGseClientServerClose = bkmonitoring.NewInt("gse_client_server_close") //服务端连接断开计数
	metricGseClientSendTimeout = bkmonitoring.NewInt("gse_client_send_timeout") //采集器发送超时次数
	metricGseClientSendRetry   = bkmonitoring.NewInt("gse_client_send_retry")   //重试的请求数
	metricGseClientSendFailed  = bkmonitoring.NewInt("gse_client_send_failed")  //发送失败的数量
	metricGseClientSendTotal   = bkmonitoring.NewInt("gse_client_send_total")   //发送成功的数量

	metricGseAgentReceived      = bkmonitoring.NewInt("gse_agent_received")
	metricGseAgentReceiveFailed = bkmonitoring.NewInt("gse_agent_receive_failed")
)

// GseClient : gse client
// used for send data and get agent info
type GseClient struct {
	socket       GseConnection
	agentInfo    AgentInfo
	quitChan     chan bool
	connectTimes uint        // 用于重连计数：达到限额后，将使用原socket进行通讯
	msgChan      chan GseMsg // msg queue
	msgQueueSize uint        // msg queue szie
	cfg          Config
}

// NewGseClient create a gse client
// host set to default gse ipc path, different from linux and windows
func NewGseClient(cfg *common.Config) (*GseClient, error) {
	// parse config
	c := defaultConfig
	err := cfg.Unpack(&c)
	if err != nil {
		logp.Err("unpack config error, %v", err)
		return nil, err
	}
	logp.Info("gse client config: %+v", c)

	cli := GseClient{
		cfg:          c,
		connectTimes: 0,
		msgQueueSize: c.MsgQueueSize,
	}
	cli.socket = NewGseConnection()
	cli.socket.SetWriteTimeout(c.WriteTimeout)
	if c.Endpoint != "" {
		cli.socket.SetHost(c.Endpoint)
	}
	return &cli, nil
}

func NewGseClientFromConfig(c Config) (*GseClient, error) {
	logp.Info("gse client config: %+v", c)

	cli := GseClient{
		cfg:          c,
		msgQueueSize: c.MsgQueueSize,
	}
	cli.socket = NewGseConnection()
	cli.socket.SetWriteTimeout(c.WriteTimeout)
	if c.Endpoint != "" {
		cli.socket.SetHost(c.Endpoint)
	}
	return &cli, nil
}

// Start : start client
// start to recv msg and get agent info
// run as goroutine
func (c *GseClient) Start() error {
	c.msgChan = make(chan GseMsg, c.msgQueueSize)
	c.quitChan = make(chan bool)

	err := c.connect()
	if err != nil {
		return err
	}

	go c.recvMsgFromAgent()
	// default request agent info evry 31s
	go c.updateAgentInfo(time.Second * 31)
	go c.msgSender()
	logp.Info("gse client start")
	return nil
}

// Close : release resources
func (c *GseClient) Close() {
	logp.Err("gse client closed")
	close(c.quitChan)
	c.socket.Close()
	return
}

// ==========================================

// GetAgentInfo : get agent info
// client will update info from gse agent every 1min
// request from agent first time when client start
func (c *GseClient) GetAgentInfo() (AgentInfo, error) {
	return c.agentInfo, nil
}

// Send : send msg to client
// will bolck when queue is full
func (c *GseClient) Send(msg GseMsg) error {
	c.msgChan <- msg
	metricGseClientReceived.Add(1)
	return nil
}

// SendWithNewConnection : send msg to client with new connection every time
func (c *GseClient) SendWithNewConnection(msg GseMsg) error {
	// new connection
	socket := NewGseConnection()
	err := socket.Dial()
	if err != nil {
		return err
	}
	defer socket.Close()

	retry := 3
	var n int
	for retry > 0 {
		n, err = socket.Write(msg.ToBytes())
		if err == nil {
			logp.Debug("gse", "send size: %d", n)
			break
		} else {
			logp.Err("gse client sendRawData failed, %v", err)
			c.reconnect()
			time.Sleep(1)
			retry--
		}
	}

	logp.Debug("gse", "send with new conneciton")
	return nil
}

// connect : connect to agent
// try to connect again several times until connected
// program will quit if failed finaly
func (c *GseClient) connect() error {
	retry := c.cfg.RetryTimes
	var err error
	for retry > 0 {
		err = c.socket.Dial()
		if err == nil {
			metricGseClientConnected.Add(1)
			logp.Info("gse client socket connected")
			return nil
		}
		metricGseClientConnectRetry.Add(1)
		logp.Err("try %d times", c.cfg.RetryTimes-retry)
		time.Sleep(c.cfg.RetryInterval)
		retry--
	}
	metricGseClientConnectFailed.Add(1)
	return err
}

// reconnect: reconnect to agent
func (c *GseClient) reconnect() {
	logp.Err("gse client reconnecting...")

	// close quitChan will stop updateAgentInfo and msgSender goroutine
	//close(c.quitChan)
	c.socket.Close()

	err := c.connect()
	if err != nil {
		logp.WTF("connect failed, program quit %v", err)
		return
	}
}

// request agent info every interval time
func (c *GseClient) updateAgentInfo(interval time.Duration) {
	logp.Info("gse client start update agent info")
	err := c.requestAgentInfo()
	if err != nil {
		logp.Err("gse client send sync cfg command failed, %v", err)
	}
	for {
		select {
		case <-time.After(interval):
			logp.Debug("gse", "send sync cfg command")
			err := c.requestAgentInfo()
			if err != nil {
				logp.Err("gse client send sync cfg command failed, error %v", err)
				continue
			}
		case <-c.quitChan:
			logp.Err("gse client updateAgentInfo quit")
			return
		}
	}
}

// msgSender : get msg from queue, send it to agent
func (c *GseClient) msgSender() {
	logp.Info("gse client start send msg")
	for {
		select {
		case msg := <-c.msgChan:
			err := c.sendRawData(msg.ToBytes())
			if err != nil {
				metricGseClientSendFailed.Add(1)
				// program quit if send error
				logp.Err("gse client send failed")
				continue
			}
			metricGseClientSendTotal.Add(1)
		case <-c.quitChan:
			logp.Err("gse client msgSender quit")
			return
		}
	}
}

// sendRawData : send binary data
func (c *GseClient) sendRawData(data []byte) error {
	retry := c.cfg.RetryTimes
	var err error
	var n int
	for retry > 0 {
		n, err = c.socket.Write(data)
		if err == nil {
			logp.Debug("gse", "send size: %d", n)
			c.onWriteSuccess()
			break
		}

		//发送重试: 根据连接状态及全局连接次数判断是否需要重连
		metricGseClientSendRetry.Add(1)
		opErrno := c.getOpErrno(err)
		isReconnect := c.isReconnectable(opErrno)
		if isReconnect {
			c.reconnect()
			c.onReconnectSuccess()
		}
		logp.Err(
			"gse client sendRawDat failed: isReconnect=>%t, connectTimes=>%d, Err=>%v",
			isReconnect, c.connectTimes, err)
		time.Sleep(c.cfg.RetryInterval)

		//如果写超时则持续写入，避免数据丢失
		if opErrno == ErrIOTimeout {
			continue
		}
		retry--
	}
	return err
}

//获取unix连接异常信息
func (c *GseClient) getOpErrno(err error) int {
	if err == syscall.EINVAL {
		return EINVAL
	}
	// 转换成*net.OpError
	opErr := (*net.OpError)(unsafe.Pointer(reflect.ValueOf(err).Pointer()))
	opError := opErr.Err.Error()
	if strings.Contains(opError, "i/o timeout") {
		return ErrIOTimeout
	} else if strings.Contains(opError, "pipe") {
		return EPIPE
	} else if strings.Contains(opError, "use of closed network connection") {
		return ErrNetClosing
	}
	return EINVAL
}

// isReconnectable： 用于写失败的重连判断
func (c *GseClient) isReconnectable(opErrno int) bool {
	//写超时使用原连接进行重试
	if opErrno == ErrIOTimeout {
		metricGseClientSendTimeout.Add(1)
		return false
	}
	//连接关闭后，直接进行重连
	if opErrno == ErrNetClosing {
		metricGseClientClose.Add(1)
		return true
	}

	//当连接次数超过重连次数限制，使用原socket进行通讯
	return c.cfg.ReconnectTimes >= c.connectTimes
}

//如果是服务端关闭，则重设连接次数
func (c *GseClient) onServerClose() {
	c.connectTimes = 0
	return
}

// 向gse agent写入成功时减少重连次数
func (c *GseClient) onReconnectSuccess() {
	c.connectTimes++
	return
}

// 向gse agent写入失败的处理操作
func (c *GseClient) onWriteSuccess() {
	if c.connectTimes > 0 {
		c.connectTimes--
	}
	return
}

// RequestAgentInfo : request agent info
func (c *GseClient) requestAgentInfo() error {
	logp.Debug("gse", "request agent info")
	msg := NewGseRequestConfMsg()
	return c.Send(msg)
}

// agentInfoMsgHandler: parse to agent info
func (c *GseClient) agentInfoMsgHandler(buf []byte) {
	var err error
	if err = json.Unmarshal(buf, &c.agentInfo); nil != err {
		logp.Err("gse client data is not json, %s", string(buf))
	}
	c.agentInfo.Hostname, err = os.Hostname()
	if err != nil {
		c.agentInfo.Hostname = ""
	}
	logp.Debug("gse", "update agent info, %+v", c.agentInfo)
}

func (c *GseClient) recvMsgFromAgent() {
	logp.Info("gse client start recv msg")
	for {
		select {
		case <-c.quitChan:
			logp.Err("gse client msgSender quit")
			return
		default:
			// read head
			headbufLen := 8 // GseLocalCommandMsg size
			headbuf := make([]byte, headbufLen)
			len, err := c.socket.Read(headbuf)

			// err handle
			if err == io.EOF {
				// socket closed by agent
				logp.Err("socket closed by remote")
				metricGseClientServerClose.Add(1)
				c.reconnect()
				c.onServerClose()
				continue
			} else if err != nil {
				metricGseAgentReceiveFailed.Add(1)
				logp.Err("gse client recv err %v", err)
				time.Sleep(time.Second)
				continue
			} else if len != headbufLen {
				metricGseAgentReceiveFailed.Add(1)
				logp.Err("gse client recv only %d bytes", len)
				continue
			}
			metricGseAgentReceived.Add(1)

			logp.Debug("gse", "recv len : %d", len)
			//logp.Debug("gse", "headbuf : %s", headbuf)

			// get type and data len
			var msg GseLocalCommandMsg
			msg.MsgType = binary.BigEndian.Uint32(headbuf[:4])
			msg.BodyLen = binary.BigEndian.Uint32(headbuf[4:])
			logp.Debug("gse", "msg type=%d, len=%d", msg.MsgType, msg.BodyLen)

			// TODO now only has GSE_TYPE_GET_CONF type
			if msg.MsgType == GSE_TYPE_GET_CONF {
				// read data
				databuf := make([]byte, msg.BodyLen)
				if _, err := c.socket.Read(databuf); nil != err && err != io.EOF {
					logp.Err("gse client read err, %v", err)
					continue
				}
				c.agentInfoMsgHandler(databuf)
			} else {
				// get other data
			}
		}
	}
	logp.Err("gse client recvMsgFromAgent quit")
}
