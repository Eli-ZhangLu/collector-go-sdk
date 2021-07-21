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

package reloader

// use signal SIGUSR1 for ipc

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/elastic/beats/libbeat/cfgfile"
	"github.com/elastic/beats/libbeat/logp"

	"github.com/TencentBlueKing/collector-go-sdk/v2/bkbeat/pidfile"
)

// NewReloader creates new Reloader instance for the given config
func NewReloader(name string, hadnler ReloadIf) *Reloader {
	return &Reloader{
		hadnler: hadnler,
		name:    name,
		done:    make(chan struct{}),
	}
}

// Run runs the reloader
func (rl *Reloader) Run(_ string) error {
	logp.Info("Config reloader started")

	// watch SIGUSR1
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGUSR1)
	go rl.signalHandler(c)
	return nil
}

// Stop stops the reloader and waits for all modules to properly stop
func (rl *Reloader) Stop() {
	close(rl.done)
}

func (rl *Reloader) signalHandler(c chan os.Signal) {
	for {
		select {
		case <-rl.done:
			logp.Info("config reloader stopped")
			return
		case s := <-c:
			logp.Info("got signal: %+v", s)
			if s == syscall.SIGUSR1 { // reload signal
				logp.Info("Reloading %s config", rl.name)

				// get new config
				c, err := cfgfile.Load("", nil)
				if err != nil {
					logp.Err("Error loading config: %s", err)
					continue
				}
				c, err = c.Child(rl.name, -1)
				if err != nil {
					logp.Err("Error loading config: %s", err)
					continue
				}

				logp.Info("reloader get config:%+v", c)

				rl.hadnler.Reload(c)
			}
		}
	}
}

// ReloadEvent send reload event
func ReloadEvent(_, pidFilePath string) error {
	fmt.Print("sending reload signal...")
	// get pid from pidfile

	pid, err := pidfile.GetPid(pidFilePath)
	if err != nil {
		fmt.Println("\033[031;1mFail\033[0m")
		return err
	}

	// send signal
	proc, err := os.FindProcess(pid)
	if err != nil {
		fmt.Println("\033[031;1mFail\033[0m")
		return err
	}
	proc.Signal(syscall.SIGUSR1)
	fmt.Println("\033[032;1mDone\033[0m")
	return nil
}
