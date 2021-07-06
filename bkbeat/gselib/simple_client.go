package gselib

import (
	"encoding/binary"
	"encoding/json"
	"io"
)

// GseSimpleClient : gse client
// used for send data and get agent info
type GseSimpleClient struct {
	socket    GseConnection
	agentInfo AgentInfo
}

// NewGseSimpleClient create a gse client
// host set to default gse ipc path, different from linux and windows
func NewGseSimpleClient() *GseSimpleClient {
	cli := GseSimpleClient{}
	cli.socket = NewGseConnection()
	return &cli
}

// Start : start client
// start to recv msg and get agent info
// run as goroutine
func (c *GseSimpleClient) Start() error {
	err := c.socket.Dial()
	if err != nil {
		return err
	}
	return nil
}

// Close : release resources
func (c *GseSimpleClient) Close() {
	c.socket.Close()
	return
}

// ==========================================
// ---- set methods should call before Start()----

// SetAgentHost : set agent host
func (c *GseSimpleClient) SetAgentHost(host string) {
	if host != "" {
		c.socket.SetHost(host)
	}
}

// ==========================================

// GetAgentInfo : get agent info
// client will update info from gse agent every 1min
// request from agent first time when client start
func (c *GseSimpleClient) GetAgentInfo() (AgentInfo, error) {
	return c.agentInfo, nil
}

// Send : send msg to client
// will bolck when queue is full
func (c *GseSimpleClient) Send(msg GseMsg) error {
	_, err := c.socket.Write(msg.ToBytes())
	return err
}

// SyncGetAgentInfo : sync request agent info
func (c *GseSimpleClient) SyncGetAgentInfo() (AgentInfo, error) {
	// request
	msg := NewGseRequestConfMsg()
	if err := c.Send(msg); err != nil {
		return AgentInfo{}, err
	}

	err := c.recvFromAgent()
	return c.agentInfo, err
}

func (c *GseSimpleClient) recvFromAgent() error {
	// read head
	headbufLen := 8 // GseLocalCommandMsg size
	headbuf := make([]byte, headbufLen)
	len, err := c.socket.Read(headbuf)
	// err handle
	if err != nil {
		return err
	} else if len != headbufLen {
		return err
	}

	// get type and data len
	var msg GseLocalCommandMsg
	msg.MsgType = binary.BigEndian.Uint32(headbuf[:4])
	msg.BodyLen = binary.BigEndian.Uint32(headbuf[4:])
	// logp.Debug("gse", "msg type=%d, len=%d", msg.msgtype, msg.bodylen)

	// now only has GSE_TYPE_GET_CONF type
	if msg.MsgType == GSE_TYPE_GET_CONF {
		// read data
		databuf := make([]byte, msg.BodyLen)
		if _, err := c.socket.Read(databuf); nil != err && err != io.EOF {
			return err
		}

		if err := json.Unmarshal(databuf, &c.agentInfo); nil != err {
			return err
		}
	} else {
		// get other data
	}
	return nil
}
