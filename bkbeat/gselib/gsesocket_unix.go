// +build linux darwin aix

package gselib

import (
	"net"
	"time"
)

const (
	// defaultGSEPath : default gse ipc path
	defaultGSEPath = "/usr/local/gse/gseagent/ipc.state.report"
	unixType       = "unix"
)

// GseLinuxConnection : gse socket struct on Linux
type GseLinuxConnection struct {
	conn         *net.UnixConn
	host         string
	netType      string
	agentInfo    AgentInfo
	writeTimeout time.Duration
}

// NewGseConnection : create a gse client
// host set to default gse ipc path, different from linux and windows
func NewGseConnection() *GseLinuxConnection {
	conn := GseLinuxConnection{
		host:    defaultGSEPath,
		netType: unixType,
	}
	return &conn
}

// Dial : connect to gse agent
func (c *GseLinuxConnection) Dial() error {
	addr := net.UnixAddr{Name: c.host, Net: unixType}
	var err error
	c.conn, err = net.DialUnix(addr.Net, nil, &addr)
	return err
}

// Close : release resources
func (c *GseLinuxConnection) Close() error {
	return c.conn.Close()
}

func (c *GseLinuxConnection) SetWriteTimeout(t time.Duration) {
	c.writeTimeout = t
}

func (c *GseLinuxConnection) Write(b []byte) (int, error) {
	if c.writeTimeout > 0 {
		err := c.conn.SetWriteDeadline(time.Now().Add(c.writeTimeout))
		if err != nil {
			return -1, err
		}
	}
	return c.conn.Write(b)
}

func (c *GseLinuxConnection) Read(b []byte) (int, error) {
	return c.conn.Read(b)
}

// SetHost : set agent host
func (c *GseLinuxConnection) SetHost(host string) {
	c.host = host
}
