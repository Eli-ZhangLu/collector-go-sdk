// Tencent is pleased to support the open source community by making
// 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.
//

package gse

import "time"

type Config struct {
	// gse client config
	RetryTimes     uint          `config:"retrytimes"`
	RetryInterval  time.Duration `config:"retryinterval"`
	Nonblock       bool          `config:"nonblock"`
	EventBufferMax int32         `config:"eventbuffermax"`
	MsgQueueSize   uint32        `config:"mqsize"`
	Endpoint       string        `config:"endpoint"`
	WriteTimeout   time.Duration `config:"writetimeout"` // unit: second

	// monitor config
	MonitorID  int32 `config:"monitorid"`  // <= 0 : disable bk monitor tag
	ResourceID int32 `config:"resourceid"` // <= 0 : disable resource report
}

var (
	defaultConfig = Config{
		MonitorID: 295,
	}
)
