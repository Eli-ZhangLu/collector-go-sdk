package gselib

import (
	"testing"
	"time"

	"github.com/elastic/beats/libbeat/common"
)

var config *common.Config

func Test_Send_NewGseDynamicMsg(t *testing.T) {
	cli, err := NewGseClient(config)
	if err != nil {
		t.Fatal(err)
	}

	err = cli.Start()
	if err != nil {
		t.Fatal(err)
	}

	m := NewGseDynamicMsg([]byte("test hc"), 1430, 0, 0)
	m.AddMeta("tlogc", "2017-01-09 21:18")
	cli.Send(m)
	time.Sleep(3 * time.Second)
	cli.Close()
}

func newOpMsg() GseMsg {
	date := time.Now().String()
	m := NewGseOpMsg([]byte(date), 1430, 0, 0, 0)
	return m
}

func Test_SendWithNewConnection(t *testing.T) {
	cli, err := NewGseClient(config)
	if err != nil {
		t.Fatal(err)
	}
	err = cli.Start()
	if err != nil {
		t.Fatal(err)
	}

	t.Log("send one op data")
	cli.SendWithNewConnection(newOpMsg())
	time.Sleep(1 * time.Second)
	t.Log("send one op data")
	cli.SendWithNewConnection(newOpMsg())
	time.Sleep(1 * time.Second)
	t.Log("send one op data")
	cli.SendWithNewConnection(newOpMsg())
	time.Sleep(1 * time.Second)
	cli.Close()
}
