package zk

import ()

// Zookeeper的数据结构
type ZK struct {
	lastZxid           int64
	xid                int32
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
