// Tencent is pleased to support the open source community by making
// 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.
//

package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
	"unsafe"

	"github.com/TencentBlueKing/collector-go-sdk/v2/bkbeat/beat"
	"github.com/TencentBlueKing/collector-go-sdk/v2/bkbeat/logp"
	"github.com/TencentBlueKing/collector-go-sdk/v2/examplebeat/config"
)

// ExampleBeat
type ExampleBeat struct {
	dataid   int32
	timer    *time.Timer
	interval time.Duration
	done     chan bool
}

func ignoreSignal(c chan os.Signal) {
	for {
		s := <-c
		logp.L.Infof("got signal: %+v", s)
		// do nothing
	}
}

func main() {
	// ignore SIGPIPE, SIGPIPE will quit process by default behavior
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGPIPE)
	go ignoreSignal(c)

	beatName := "examplebeat"
	version := "V1.0"
	var err error
	var localconfig *beat.Config

	localconfig, err = beat.Init(beatName, version)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	cfg := config.DefaultConfig
	err = localconfig.Unpack(&cfg)

	// 指定发送数据大小
	var msgArr []byte
	var msg string
	if cfg.DataSize > 0 {
		msgArr = make([]byte, cfg.DataSize)
		memsetRepeat(msgArr, 65)
		msg = *(*string)(unsafe.Pointer(&msgArr))
	} else {
		msg = ""
	}
	// 指定发送时间间隔
	timer := time.NewTicker(cfg.Interval)
	for {
		select {
		case <-timer.C:
			event := beat.MapStr{
				"dataid": cfg.DataID,
			}
			if len(msg) > 0 {
				event["msg"] = msg
			}
			beat.Send(event)

		// 采集器被停止
		case <-beat.Done:
			return

		// 配置重加载
		case <-beat.ReloadChan:
			localconfig = beat.GetConfig()
			cfg = config.DefaultConfig
			err = localconfig.Unpack(&cfg)
			if err != nil {
				fmt.Println(err.Error())
				continue
			}
			timer = time.NewTicker(cfg.Interval)
			if cfg.DataSize > 0 {
				msgArr = make([]byte, cfg.DataSize)
				memsetRepeat(msgArr, 65)
				msg = *(*string)(unsafe.Pointer(&msgArr))
			} else {
				msg = ""
			}
		}
	}
}

func memsetRepeat(a []byte, v byte) {
	if len(a) == 0 {
		return
	}
	a[0] = v
	for bp := 1; bp < len(a); bp *= 2 {
		copy(a[bp:], a[:bp])
	}
}
