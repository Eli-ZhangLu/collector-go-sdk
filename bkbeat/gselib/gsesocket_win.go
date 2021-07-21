// Tencent is pleased to support the open source community by making
// 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.
//

// +build windows

package gselib

import (
	"log"
	"net"
	"time"
)

const (
	// defaultGsePath : default gse ipc path
	defaultGSEPath = "127.0.0.1:47000"
	tcpType        = "tcp"
)

// GseWindowsConnection : gse socket struct on Linux
type GseWindowsConnection struct {
	conn         *net.TCPConn
	host         string
	netType      string
	agentInfo    AgentInfo
	writeTimeout time.Duration
}

// NewGseConnection : create a gse client
// host set to default gse ipc path, different from linux and windows
func NewGseConnection() *GseWindowsConnection {
	conn := GseWindowsConnection{
		host:    defaultGSEPath,
		netType: tcpType,
	}
	return &conn
}

// Dial : connect to gse agent
func (c *GseWindowsConnection) Dial() error {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", c.host)
	if err != nil {
		log.Println("Err: ResolveTCPAddr error")
		return err
	}
	c.conn, err = net.DialTCP(c.netType, nil, tcpAddr)
	if err != nil {
		log.Println("Err: DialTCP error")
		return err
	}
	return nil
}

// Close : release resources
func (c *GseWindowsConnection) Close() error {
	return c.conn.Close()
}

func (c *GseWindowsConnection) SetWriteTimeout(t time.Duration) {
	c.writeTimeout = t
}

func (c *GseWindowsConnection) Write(b []byte) (int, error) {
	if c.writeTimeout > 0 {
		err := c.conn.SetWriteDeadline(time.Now().Add(c.writeTimeout))
		if err != nil {
			return -1, err
		}
	}
	return c.conn.Write(b)
}

func (c *GseWindowsConnection) Read(b []byte) (int, error) {
	return c.conn.Read(b)
}

// SetHost : set agent host
func (c *GseWindowsConnection) SetHost(host string) {
	c.host = host
}
