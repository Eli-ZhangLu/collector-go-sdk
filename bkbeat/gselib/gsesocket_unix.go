// Tencent is pleased to support the open source community by making
// 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.
//

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
