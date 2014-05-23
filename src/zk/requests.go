package zk

import ()

type requestHeader struct {
	Xid    int32
	Opcode int32
}

// 基本请求结构
type request struct {
	xid    int32
	opcode int32

	pkt        interface{}
	recvStruct interface{}
	recvChan   chan response

	recvFunc func(*request, *responseHeader, err)
}

type closeRequest struct{}

type connectRequest struct {
	ProtocolVersion int32
	LastZxidSeen    int64
	TimeOut         int32
	SessionId       int64
	Passwd          []byte
}

type getDataRequest struct {
	Path  string
	Watch bool
}

type createRequest struct {
	Path  string
	Data  []byte
	Acl   []ACL
	Flags int32
}
