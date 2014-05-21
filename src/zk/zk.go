package zk

import (
	"fmt"
	"net"
	//"strconv"
	//"strings"
	"encoding/binary"
	"io"
	"time"
)

const (
	bufferSize      = 1536 * 1024 // 缓冲区大小，单位是字节
	protocolVersion = 0
)

var (
	// 第一次请求时，客户端发送16位空字节数组到服务器端
	emptyPassword = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
)

// Zookeeper的数据结构
type ZK struct {
	lastZxid           int64
	servers            []string      // 服务器地址集合
	connectTimeout     time.Duration // 连接超时
	sessionId          int64         // 会话Id
	sessionTimeout     int32         // 会话超时
	conn               net.Conn      // 标准库里的TCP连接结构体
	password           []byte        // 密码
	connectRequest     interface{}   // 连接请求
	getRequest         interface{}   // 获取数据请求
	getChildrenRequest interface{}   // 获取所有子节点请求
	createRequest      interface{}   // 新建节点请求
	modifyRequest      interface{}   // 修改节点请求
	deleteRequest      interface{}   // 删除节点请求
}

// Znode的数据结构
type ZNode struct {
	Czxid           int64 // 节点被创建的Zxid值
	Ctime           int64 // 节点被创建的时间
	Mzxid           int64 // 节点被修改的Zxid值
	Mtime           int64 // 节点被修改的时间
	DataVersion     int32 // 节点数据修改的版本号
	ChildrenVersion int32 // 节点的子节点被修改的版本号
	AclVersion      int32 // 节点的ACL被修改的版本号
	DataLength      int32 // 节点数据长度
	NumChildren     int32 // 节点的子节点数量
	EphemeralOwner  int64 // 如果不是临时节点，值为0，否则为节点拥有者的会话Id
}

// 基本请求结构
type request struct {
	xid    int32
	opcode int32
}

// 基本响应结构
type response struct {
	zxid int64
	err  error
}

// 连接到服务器
func Connect(servers []string, connectTimeout time.Duration) (ZK, error) {
	fmt.Println("connecting...")
	zk := ZK{
		servers:        servers,
		connectTimeout: connectTimeout,
		sessionId:      0,
		sessionTimeout: 30000,
		conn:           nil,
		password:       emptyPassword,
	}
	var conn net.Conn
	var err error = nil
	// 尝试所有IP，一有成功连接的，马上跳出
	for _, serverip := range servers {
		conn, err = net.DialTimeout("tcp", serverip, connectTimeout)
		if err == nil {
			zk.conn = conn
			break
		}
	}
	// 如果连接成功，则进行认证
	if zk.conn != nil {
		buf := make([]byte, 256)
		zk.connectRequest = &connectRequest{
			ProtocolVersion: protocolVersion,
			LastZxidSeen:    zk.lastZxid,
			TimeOut:         zk.sessionTimeout,
			SessionId:       zk.sessionId,
			Passwd:          zk.password,
		}
		var n int
		n, err = encodePacket(buf, zk.connectRequest)
		binary.BigEndian.PutUint32(buf[:4], uint32(n))

		// 发送数据到服务器
		_, err = zk.conn.Write(buf[:n+4])

		_, err = io.ReadFull(zk.conn, buf[:4])

	}
	return zk, err
}

// 断开与服务器的连接
func (zk *ZK) Close() {

}

// 获取节点数据
func (zk *ZK) Get(path string) []byte {
	ret := []byte("test")

	return ret
}

// 新建一个节点
func (zk *ZK) Create(path string, data []byte) {

}
