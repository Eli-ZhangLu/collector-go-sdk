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
	"flag"
	"fmt"
	"os"

	"github.com/TencentBlueKing/collector-go-sdk/v2/bkbeat/storage"
)

var path string
var key string

func init() {
	flag.StringVar(&path, "path", "", "db path")
	flag.StringVar(&key, "key", "", "query key")
}

func main() {
	flag.Parse()
	if path == "" || key == "" {
		flag.PrintDefaults()
		os.Exit(0)
	}
	err := storage.Init(path, nil)
	if err != nil {
		panic(err)
	}
	v, err := storage.Get(key)
	if err == storage.ErrNotFound {
		fmt.Println("not found.")
		os.Exit(0)
	}
	if err != nil {
		panic(err)
	}
	fmt.Println(v)
}
