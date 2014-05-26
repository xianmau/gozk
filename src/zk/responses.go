package zk

import ()

// 响应包头
type responseHeader struct {
	Xid  int32 // -2表示ping的响应，-4表示auth的响应，-1表示本次响应有watch事件发生
	Zxid int64
	Err  int32
}

// 响应结构
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

// 关闭请求响应
type closeResponse struct{}

// Ping请求响应
type pingResponse struct{}

// 判断存在请求响应
type existsResponse struct {
	Stat Stat
}

// 添加请求响应
type createResponse struct {
	Path       string
	Data       []byte
	Acl        []ACL
	SessionACL []byte
}

// 获取子节点请求响应
type getChildrenResponse struct {
	Children []string
}

// 获取节点数据请求响应
type getDataResponse struct {
	Data []byte
	Stat Stat
}

// 设置节点数据请求响应
type setDataResponse struct {
	Stat Stat
}

// 删除请求响应
type deleteResponse struct{}
