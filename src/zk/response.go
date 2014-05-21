package zk

import ()

type connectResponse struct {
	ProtocolVersion int32
	TimeOut         int32
	SessionId       int64
	Passwd          []byte
}
