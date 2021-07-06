package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/TencentBlueKing/collector-go-sdk/bkbeat/storage"
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
