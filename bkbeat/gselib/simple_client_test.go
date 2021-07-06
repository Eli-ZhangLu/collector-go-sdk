package gselib

import (
	"testing"
	"time"
)

func Test_GseSimpleClient(t *testing.T) {
	cli := NewGseSimpleClient()
	cli.SetAgentHost(MockAddress)
	err := cli.Start()
	if err != nil {
		t.Fatal(err)
	}

	info := AgentInfo{}
	// get agent info
	go func() {
		info, err = cli.SyncGetAgentInfo()
		if err != nil {
			t.Fatal(err)
		}
	}()

	// timewait to get agent info
	time.Sleep(1 * time.Second)

	if info.IP == "" {
		t.Fatal("request agent info timeout")
	}

	if info.Bizid != 1 {
		t.Fatal("get companyid error")
	}
	if info.Cloudid != 2 {
		t.Fatal("get plat_id error")
	}
	if info.IP != "10.0.0.1" {
		t.Fatal("get ip error")
	}

	cli.Close()

	// send msg
	md := NewGseDynamicMsg([]byte("test dynamic"), 1430, 0, 0)
	md.AddMeta("tlogc", "2017-01-09 21:18")
	cli.Send(md)

	// send msg
	mc := NewGseCommonMsg([]byte("test common"), 1430, 0, 0, 0)
	cli.Send(mc)
}
