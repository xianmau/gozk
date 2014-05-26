package zk

import (
	"net"
	"sync"
	"time"
)

// Zookeeper的数据结构
type ZK struct {
	lastZxid       int64         // 最后修改的节点
	xid            int32         // 用来标记请求的版本
	servers        []string      // 服务器地址集合
	serversIndex   int           // 当前连接的服务器序号
	conn           net.Conn      // 标准库里的TCP连接结构体
	connectTimeout time.Duration // 连接超时
	sessionId      int64         // 会话Id
	sessionTimeout int32         // 会话超时
	password       []byte        // 密码
	state          int32         // 连接状态

	heartbeatInterval time.Duration // 心跳时间，一般为接收超时的一半
	recvTimeout       time.Duration // 接收超时，一般为心跳的两倍

	shouldQuit   chan bool          // 是否退出
	sendChan     chan *request      // 请求队列，客户端把要发送的请求丢到这里，可以实现同步
	requests     map[int32]*request // 正在准备的请求映射，在请求结构中有Xid，可用来对应哪个请求
	requestsLock sync.Mutex         // 请求锁

}

// Znode状态的数据结构
type Stat struct {
	Czxid          int64 // 节点被创建的Zxid值
	Mzxid          int64 // 节点被修改的Zxid值
	Ctime          int64 // 节点被创建的时间
	Mtime          int64 // 节点被修改的时间
	Version        int32 // 节点数据修改的版本号
	CVersion       int32 // 节点的子节点被修改的版本号
	AVersion       int32 // 节点的ACL被修改的版本号
	EphemeralOwner int64 // 如果不是临时节点，值为0，否则为节点拥有者的会话Id
	DataLength     int32 // 节点数据长度
	NumChildren    int32 // 节点的子节点数量
	Pzxid          int64 // 最后修改子节点
}
