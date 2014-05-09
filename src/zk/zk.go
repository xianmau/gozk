package zk

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

const (
	bufferSize      = 1536 * 1024
	eventChanSize   = 6
	sendChanSize    = 16
	protectedPrefix = "_c_"
)

type Conn struct {
	lastZxid  int64
	sessionId int64
	//state State
	xid     int32
	timeout int32
	passwd  []byte

	dialer         Dialer
	servers        []string
	serverIndex    int
	conn           net.Conn
	eventChan      chan Event
	shouldQuit     chan bool
	pingInterval   time.Duration
	recvTimeout    time.Duration
	connectTimeout time.Duration
}
type Event struct {
	Type  EventType
	State State
	Path  string
}
type Dialer func(network string, address string, timeout time.Duration) (net.Conn, error)

func Connect(servers []string, recvTimeout time.Duration) (*Conn, <-chan Event, error) {
	fmt.Println("Call [ConnectWithDialer]")
	return ConnectWithDialer(servers, recvTimeout, nil)
}

func ConnectWithDialer(servers []string, recvTimeout time.Duration, dialer Dialer) (*Conn, <-chan Event, error) {
	fmt.Println("Exec [ConnectWithDialer]")
	for i, addr := range servers {
		if !strings.Contains(addr, ":") {
			servers[i] = addr + ":" + strconv.Itoa(DefaultPort)
		}
	}

	ec := make(chan Event, eventChanSize)
	if dialer == nil {
		dialer = net.DialTimeout
	}

	return nil, nil, nil
}
