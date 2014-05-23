package zk

import ()

// 响应包头
type responseHeader struct {
	Xid  int32
	Zxid int64
	Err  int32
}

// 响应结构
type response struct {
	zxid int64
	err  error
}

// 连接请求响应
type connectResponse struct {
	ProtocolVersion int32
	TimeOut         int32
	SessionId       int64
	Passwd          []byte
}

// 获取节点数据请求响应
type getDataResponse struct {
	Data []byte
	Stat Stat
}
