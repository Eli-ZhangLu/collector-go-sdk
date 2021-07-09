package main

import (
	"flag"

	"github.com/TencentBlueKing/collector-go-sdk/v2/bkbeat/gselib"
)

func main() {
	flag.Parse()
	gselib.StartMockAgent()
}
