package zk

import ()

// 请求包头
type requestHeader struct {
	Xid    int32
	Opcode int32
}

// 请求结构
type request struct {
	xid    int32
	opcode int32

	pkt        interface{}
	recvStruct interface{}
	recvChan   chan response

	recvFunc func(*request, *responseHeader, error)
}

// 关闭连接请求
type closeRequest struct{}

type connectRequest struct {
	ProtocolVersion int32
	LastZxidSeen    int64
	TimeOut         int32
	SessionId       int64
	Passwd          []byte
}

// 获取节点数据请求
type getDataRequest struct {
	Path  string
	Watch bool
}

// 创建新节点请求
type createRequest struct {
	Path  string
	Data  []byte
	Acl   []ACL
	Flags int32
}
