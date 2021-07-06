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
