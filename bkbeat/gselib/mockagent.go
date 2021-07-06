package gselib

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

var fixInfo []byte
var sum int
var count int
var verbose = flag.Bool("verbose", false, "show message body")

func makeData() {
	var info = `{"bizid":1,"cloudid":2,"ip":"10.0.0.1"}`
	var infoByte = []byte(info)
	head := GseLocalCommandMsg{
		10,
		uint32(len(infoByte)),
	}
	buffer := new(bytes.Buffer)
	binary.Write(buffer, binary.BigEndian, head)
	binary.Write(buffer, binary.LittleEndian, infoByte)
	fixInfo = buffer.Bytes()
}

// handleConnection only recv data
func handleConnection(conn net.Conn) {
	for {
		// read head
		headbufLen := 24 // GseCommonMsgHead size
		headbuf := make([]byte, headbufLen)
		n, err := conn.Read(headbuf)

		// err handle
		if err == io.EOF {
			// socket closed by agent
			fmt.Println(err)
			return
		} else if err != nil {
			fmt.Println("err:", err)
			return
		} else if n != headbufLen {
			fmt.Println("want 24 !=", n)
			return
		}
		sum += n

		// get type and data len
		msgType := binary.BigEndian.Uint32(headbuf[:4])
		bodyLen := binary.BigEndian.Uint32(headbuf[12:])
		buf := make([]byte, bodyLen)
		if bodyLen > 0 {
			n, err = conn.Read(buf)
			if nil != err && err != io.EOF {
				fmt.Println("err:", err)
				return
			}
			fmt.Printf("Time: %s \t Size: %d\n", time.Now().String(), n)
			if *verbose {
				fmt.Println(string(buf))
			}
			sum += n
		}

		count += 1
		fmt.Println("recv msg type=", msgType)
		if msgType == 10 {
			// send agent info
			fmt.Println("return agent info")
			conn.Write(fixInfo)
		} else {
			// get other data
		}
	}
}

const (
	BYTE = 1.0 << (10 * iota)
	KILOBYTE
	MEGABYTE
	GIGABYTE
)

func byteSize(bytes int) string {
	unit := ""
	value := float32(bytes)

	switch {
	case bytes >= GIGABYTE:
		unit = "G"
		value = value / GIGABYTE
	case bytes >= MEGABYTE:
		unit = "M"
		value = value / MEGABYTE
	case bytes >= KILOBYTE:
		unit = "K"
		value = value / KILOBYTE
	case bytes >= BYTE:
		unit = "B"
	case bytes == 0:
		return "0"
	}

	stringValue := fmt.Sprintf("%.1f", value)
	stringValue = strings.TrimSuffix(stringValue, ".0")
	return fmt.Sprintf("%s%s", stringValue, unit)
}

// start only one mock agent
var once sync.Once

func StartMockAgent() {
	once.Do(start)
}

func start() {
	makeData()

	os.Remove(MockAddress)
	ln, err := net.Listen(MockNetwork, MockAddress)
	if err != nil {
		panic(err)
	}

	// statistic
	go func() {
		lastSize := 0
		lastCout := 0
		timer := time.NewTicker(1 * time.Second)
		for {
			select {
			case <-timer.C:
				interval := sum - lastSize
				fmt.Println(count-lastCout, byteSize(interval))
				lastSize = sum
				lastCout = count
			}
		}
	}()

	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
		}
		go handleConnection(conn)
	}
}
