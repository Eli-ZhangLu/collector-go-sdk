package main

import (
	"flag"

	"github.com/TencentBlueKing/collector-go-sdk/bkbeat/gselib"
)

func main() {
	flag.Parse()
	gselib.StartMockAgent()
}
