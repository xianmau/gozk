package zk

import ()

type responseHeader struct {
	Xid  int32
	Zxid int64
	Err  ErrCode
}

// 基本响应结构
type response struct {
	zxid int64
	err  error
}

type connectResponse struct {
	ProtocolVersion int32
	TimeOut         int32
	SessionId       int64
	Passwd          []byte
}

type getDataResponse struct {
	Data []byte
	Stat Stat
}
