package zk

/*----------------------------------------------
 *
 * 响应协议
 *
 ----------------------------------------------*/

// Connect
type connectResponse struct {
	protocolversion int32
	timeout         int32
	sessionid       int64
	password        []byte
}

func decodeConnectResponse(buf []byte, res *connectResponse) {
	res.protocolversion = BytesToInt32(buf[0:])
	res.timeout = BytesToInt32(buf[4:])
	res.sessionid = BytesToInt64(buf[8:])
	res.password = buf[16+4 : 16+4+16]
}

// Close
type closeResponse struct {
	xid     int32
	zxid    int64
	errcode int32
}

func decodeCloseResponse(buf []byte, res *closeResponse) {
	res.xid = BytesToInt32(buf[0:])
	res.zxid = BytesToInt64(buf[4:])
	res.errcode = BytesToInt32(buf[12:])
}

// Ping
type pingResponse struct {
	xid     int32
	zxid    int64
	errcode int32
}

func decodePingResponse(buf []byte, res *pingResponse) {
	res.xid = BytesToInt32(buf[0:])
	res.zxid = BytesToInt64(buf[4:])
	res.errcode = BytesToInt32(buf[12:])
}

// Exist
type existResponse struct {
	xid     int32
	zxid    int64
	errcode int32
	// 这里还有节点的状态，但目前不关心
}

func decodeExistResponse(buf []byte, res *existResponse) {
	res.xid = BytesToInt32(buf[0:])
	res.zxid = BytesToInt64(buf[4:])
	res.errcode = BytesToInt32(buf[12:])
}

// Get
type getResponse struct {
	xid     int32
	zxid    int64
	errcode int32
	data    []byte
	// 这里还有节点的状态，但目前不关心
}

func decodeGetResponse(buf []byte, res *getResponse) {
	byte_cnt := BytesToInt32(buf[0:])
	buf = buf[4:]
	res.xid = BytesToInt32(buf[0:])
	res.zxid = BytesToInt64(buf[4:])
	res.errcode = BytesToInt32(buf[12:])
	if byte_cnt > 16 {
		length := BytesToInt32(buf[16:])
		res.data = buf[20 : 20+length]
	}
}

// Children
type childrenResponse struct {
	xid      int32
	zxid     int64
	errcode  int32
	children []string
}

func decodeChildrenResponse(buf []byte, res *childrenResponse) {
	byte_cnt := BytesToInt32(buf[0:])
	buf = buf[4:]
	res.xid = BytesToInt32(buf[0:])
	res.zxid = BytesToInt64(buf[4:])
	res.errcode = BytesToInt32(buf[12:])
	if byte_cnt > 16 {
		children_cnt := BytesToInt32(buf[16:])
		var n int32 = 20
		for ; children_cnt > 0; children_cnt-- {
			child_len := BytesToInt32(buf[n:])
			n += 4
			res.children = append(res.children, string(buf[n:n+child_len]))
			n += child_len
		}
	}
}

// Create
type createResponse struct {
	xid     int32
	zxid    int64
	errcode int32
	// 这里还有节点的路径，但目前不关心
}

func decodeCreateResponse(buf []byte, res *createResponse) {
	res.xid = BytesToInt32(buf[0:])
	res.zxid = BytesToInt64(buf[4:])
	res.errcode = BytesToInt32(buf[12:])
}

// Set
type setResponse struct {
	xid     int32
	zxid    int64
	errcode int32
	// 这里还有节点的状态，但目前不关心
}

func decodeSetResponse(buf []byte, res *setResponse) {
	res.xid = BytesToInt32(buf[0:])
	res.zxid = BytesToInt64(buf[4:])
	res.errcode = BytesToInt32(buf[12:])
}

// Delete
type deleteResponse struct {
	xid     int32
	zxid    int64
	errcode int32
}

func decodeDeleteResponse(buf []byte, res *deleteResponse) {
	res.xid = BytesToInt32(buf[0:])
	res.zxid = BytesToInt64(buf[4:])
	res.errcode = BytesToInt32(buf[12:])
}
