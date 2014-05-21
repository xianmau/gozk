package zk

import ()

type closeRequest struct{}

type connectRequest struct {
	ProtocolVersion int32
	LastZxidSeen    int64
	TimeOut         int32
	SessionId       int64
	Passwd          []byte
}

type createRequest struct {
	Path  string
	Data  []byte
	Acl   []ACL
	Flags int32
}
