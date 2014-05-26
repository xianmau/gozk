package zk

import ()

// 响应包头
type responseHeader struct {
	Xid  int32 // -2表示ping的响应，-4表示auth的响应，-1表示本次响应有watch事件发生
	Zxid int64
	Err  int32
}

// 响应结构，ZooKeeper Transaction Id
type response struct {
	zxid int64 // 每次请求对应一个唯一的zxid，如果zxid a < zxid b ，则可以保证a一定发生在b之前
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
