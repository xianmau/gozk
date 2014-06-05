package zk

import (
	"net"
	"sync"
	"sync/atomic"
)

const (
	DefaultPort    = 2181     // 默认端口号
	RecvTimeout    = 1        // 接收消息超时，单位：秒
	SessionTimeout = 4000     // 客户端会话超时，单位：毫秒
	PingInterval   = 2000     // Ping超时,单位：毫秒
	BufferSize     = 2 * 1024 // 1K
	SentChanSize   = 16       // 发送请求队列大小
	RecvChanSize   = 16       // 接收响应队列大小
)

type ZkCli struct {
	xid             int32              // 请求编号，用于映射请求响应
	reqMap          map[int32]*request // 请求响应映射
	reqLock         sync.Mutex         //请求锁，在操作请求映射可能需要加锁
	protocolversion int32
	sessiontimeout  int32
	sessionid       int64
	password        []byte
	conn            *net.TCPConn
	state           int32
	sentchan        chan *request
}

type request struct {
	xid       int32           //
	opcode    int32           //
	reqbuf    []byte          // 用于直接发送的字节数组
	resheader *responseHeader //
	resbuf    []byte          // 直接接收到的字节数组
	err       error           // 错误信息
	done      chan bool       // 是否处理完成
}

// API：新建一个实例
func New() *ZkCli {
	zkCli := ZkCli{
		xid:             0,
		reqMap:          make(map[int32]*request),
		protocolversion: 0,
		sessiontimeout:  0,
		sessionid:       0,
		password:        []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		conn:            nil,
		state:           stateDisconnect,
		sentchan:        make(chan *request, SentChanSize),
	}
	return &zkCli
}

func (zkCli *ZkCli) getNextXid() int32 {
	return atomic.AddInt32(&zkCli.xid, 1)
}
