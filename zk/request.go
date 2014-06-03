package zk

/*----------------------------------------------
 *
 * 请求协议
 *
 ----------------------------------------------*/

// Connect
type connectRequest struct {
	protocolversion int32
	lastzxidseen    int64
	timeout         int32
	sessionid       int64
	password        []byte
}

func encodeConnectRequest(buf []byte, req *connectRequest) int32 {
	Int32ToBytes(buf[4:], req.protocolversion)
	Int64ToBytes(buf[8:], req.lastzxidseen)
	Int32ToBytes(buf[16:], req.timeout)
	Int64ToBytes(buf[20:], req.sessionid)
	Int32ToBytes(buf[28:], 16)
	copy(buf[32:], req.password)
	Int32ToBytes(buf[0:], 44)
	return 48
}

// Close
type closeRequest struct {
	xid    int32
	opcode int32
}

func encodeCloseRequest(buf []byte, req *closeRequest) int32 {
	Int32ToBytes(buf[4:], req.xid)
	Int32ToBytes(buf[8:], req.opcode)
	Int32ToBytes(buf[0:], 8)
	return 12
}

// Ping
type pingRequest struct {
	xid    int32
	opcode int32
}

func encodePingRequest(buf []byte, req *pingRequest) int32 {
	Int32ToBytes(buf[4:], req.xid)
	Int32ToBytes(buf[8:], req.opcode)
	Int32ToBytes(buf[0:], 8)
	return 12
}

// Exist
type existRequest struct {
	xid    int32
	opcode int32
	path   string
	watch  bool
}

func encodeExistRequest(buf []byte, req *existRequest) int32 {
	path_len := int32(len(req.path))
	Int32ToBytes(buf[4:], req.xid)
	Int32ToBytes(buf[8:], req.opcode)
	Int32ToBytes(buf[12:], path_len)
	copy(buf[16:], []byte(req.path))
	BoolToBytes(buf[16+path_len:], req.watch)
	Int32ToBytes(buf[0:], 13+path_len)
	return 17 + path_len
}

// Get
type getRequest struct {
	xid    int32
	opcode int32
	path   string
	watch  bool
}

func encodeGetRequest(buf []byte, req *getRequest) int32 {
	path_len := int32(len(req.path))
	Int32ToBytes(buf[4:], req.xid)
	Int32ToBytes(buf[8:], req.opcode)
	Int32ToBytes(buf[12:], path_len)
	copy(buf[16:], []byte(req.path))
	BoolToBytes(buf[16+path_len:], req.watch)
	Int32ToBytes(buf[0:], 13+path_len)
	return 17 + path_len
}

// Children
type childrenRequest struct {
	xid    int32
	opcode int32
	path   string
	watch  bool
}

func encodeChildrenRequest(buf []byte, req *childrenRequest) int32 {
	path_len := int32(len(req.path))
	Int32ToBytes(buf[4:], req.xid)
	Int32ToBytes(buf[8:], req.opcode)
	Int32ToBytes(buf[12:], path_len)
	copy(buf[16:], []byte(req.path))
	BoolToBytes(buf[16+path_len:], req.watch)
	Int32ToBytes(buf[0:], 13+path_len)
	return 17 + path_len
}

// Create
type createRequest struct {
	xid    int32
	opcode int32
	path   string
	data   []byte
	acl    []ACL
	flags  int32
}

func encodeCreateRequest(buf []byte, req *createRequest) int32 {
	path_len := len(req.path)
	data_len := len(req.data)
	acl_cnt := len(req.acl)
	n := 4
	Int32ToBytes(buf[n:], req.xid)
	n += 4
	Int32ToBytes(buf[n:], req.opcode)
	n += 4
	Int32ToBytes(buf[n:], int32(path_len))
	n += 4
	copy(buf[n:], []byte(req.path))
	n += path_len
	Int32ToBytes(buf[n:], int32(data_len))
	n += 4
	copy(buf[n:], req.data)
	n += data_len
	Int32ToBytes(buf[n:], int32(acl_cnt))
	n += 4
	for i := 0; i < acl_cnt; i++ {
		Int32ToBytes(buf[n:], req.acl[i].Perms)
		n += 4
		acl_scheme_len := len(req.acl[i].Scheme)
		Int32ToBytes(buf[n:], int32(acl_scheme_len))
		n += 4
		copy(buf[n:], []byte(req.acl[i].Scheme))
		n += acl_scheme_len
		acl_id_len := len(req.acl[i].Id)
		Int32ToBytes(buf[n:], int32(acl_id_len))
		n += 4
		copy(buf[n:], []byte(req.acl[i].Id))
		n += acl_id_len
	}
	Int32ToBytes(buf[n:], req.flags)
	Int32ToBytes(buf[0:], int32(n))
	return int32(n + 4)
}

// Set
type setRequest struct {
	xid     int32
	opcode  int32
	path    string
	data    []byte
	version int32
}

func encodeSetRequest(buf []byte, req *setRequest) int32 {
	path_len := len(req.path)
	data_len := len(req.data)
	n := 4
	Int32ToBytes(buf[n:], req.xid)
	n += 4
	Int32ToBytes(buf[n:], req.opcode)
	n += 4
	Int32ToBytes(buf[n:], int32(path_len))
	n += 4
	copy(buf[n:], []byte(req.path))
	n += path_len
	Int32ToBytes(buf[n:], int32(data_len))
	n += 4
	copy(buf[n:], req.data)
	n += data_len
	Int32ToBytes(buf[n:], req.version)
	Int32ToBytes(buf[0:], int32(n))
	return int32(n + 4)
}

// Delete
type deleteRequest struct {
	xid     int32
	opcode  int32
	path    string
	version int32
}

func encodeDeleteRequest(buf []byte, req *deleteRequest) int32 {
	path_len := len(req.path)
	n := 4
	Int32ToBytes(buf[n:], req.xid)
	n += 4
	Int32ToBytes(buf[n:], req.opcode)
	n += 4
	Int32ToBytes(buf[n:], int32(path_len))
	n += 4
	copy(buf[n:], []byte(req.path))
	n += path_len
	Int32ToBytes(buf[n:], req.version)
	Int32ToBytes(buf[0:], int32(n))
	return int32(n + 4)
}
