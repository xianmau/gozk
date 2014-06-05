package zk

import ()

type requestHeader struct {
	xid    int32
	opcode int32
}

func encodeRequestHeader(buf []byte, req *requestHeader) int32 {
	Int32ToBytes(buf[4:], req.xid)
	Int32ToBytes(buf[8:], req.opcode)
	Int32ToBytes(buf[0:], 8)
	return 12
}

type responseHeader struct {
	xid     int32
	zxid    int64
	errcode int32
}

func decodeResponseHeader(buf []byte, res *responseHeader) {
	res.xid = BytesToInt32(buf)
	res.zxid = BytesToInt64(buf[4:])
	res.errcode = BytesToInt32(buf[12:])
}
