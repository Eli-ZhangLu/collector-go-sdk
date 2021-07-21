// Tencent is pleased to support the open source community by making
// 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.
//

package gselib

import (
	"testing"
)

func Test_NewGseCommonMsg(t *testing.T) {
	msg := NewGseCommonMsg([]byte("test common"), 0, 0, 0, 0)
	t.Log(msg.ToBytes())
}

func Test_NewGseDynamicMsg(t *testing.T) {
	msg := NewGseDynamicMsg([]byte("test dynamic"), 0, 0, 0)
	msg.AddMeta("test_key", "test_value")
	t.Log(msg.ToBytes())
}

func Test_NewGseOpMsg(t *testing.T) {
	msg := NewGseOpMsg([]byte("test"), 0, 0, 0, 0)
	msg.ToBytes()
	t.Log(msg.ToBytes())
}

func Test_NewGseRequestConfMsg(t *testing.T) {
	msg := NewGseRequestConfMsg()
	msg.ToBytes()
	t.Log(msg.ToBytes())
}
