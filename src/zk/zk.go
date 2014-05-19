package zk

import (
	//"fmt"
	"net"
	//"strconv"
	//"strings"
	"time"
)

const (
	bufferSize      = 1536 * 1024
	eventChanSize   = 6
	sendChanSize    = 16
	protectedPrefix = "_c_"
)

var (
	// 第一次请求时，客户端发送16位空字节数组到服务器端
	emptyPassword = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
)

type ZK struct {
	// 服务器地址集合
	servers []string
	// 连接超时
	connectTimeout time.Duration
	// 会话Id
	sessionId int64
	// 会话超时
	sessionTimeout time.Duration
	// 标准库里的TCP连接结构体
	conn net.Conn
	// 密码
	password []byte

	//
	connectRequest interface{}
	//
	getRequest interface{}
	//
	getChildrenRequest interface{}
	//
	createRequest interface{}
	//
	updateRequest interface{}
	//
	deleteRequest interface{}
}

// 连接到服务器
func Connect(servers []string, connectTimeout time.Duration) (ZK, error) {
	zk := ZK{
		servers:        servers,
		connectTimeout: connectTimeout,
		sessionId:      0,
		sessionTimeout: 1000000,
		conn:           nil,
		password:       emptyPassword,
	}
	var err error = nil
	for _, serverip := range servers {
		conn, err := net.DialTimeout("tcp", serverip, connectTimeout)
		if err == nil {
			zk.conn = conn
			break
		}
	}
	return zk, err
}

// 获取节点数据
func (zk *ZK) Get(path string) []byte {
	ret := []byte("test")

	return ret
}

// 新建一个节点
func (zk *ZK) Create(path string, data []byte) {

}
