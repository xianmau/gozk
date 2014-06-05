package zk

import ()

type createRequest struct {
	xid    int32
	opcode int32
	path   string
	data   []byte
	acl    []ACL
	flags  int32
}

func encodeCreateRequest(buf []byte, req *createRequest) int32 {
	path := []byte(req.path)
	path_len := int32(len(path))
	data_len := int32(len(req.data))
	acl_cnt := len(req.acl)
	n := 4
	Int32ToBytes(buf[n:], req.xid)
	n += 4
	Int32ToBytes(buf[n:], req.opcode)
	n += 4
	Int32ToBytes(buf[n:], path_len)
	n += 4
	copy(buf[n:], path)
	n += int(path_len)
	Int32ToBytes(buf[n:], data_len)
	n += 4
	copy(buf[n:], req.data)
	n += int(data_len)
	Int32ToBytes(buf[n:], int32(acl_cnt))
	n += 4
	for i := 0; i < acl_cnt; i++ {
		Int32ToBytes(buf[n:], req.acl[i].Perms)
		n += 4
		acl_scheme := []byte(req.acl[i].Scheme)
		acl_scheme_len := len(acl_scheme)
		Int32ToBytes(buf[n:], int32(acl_scheme_len))
		n += 4
		copy(buf[n:], acl_scheme)
		n += acl_scheme_len
		acl_id := []byte(req.acl[i].Id)
		acl_id_len := len(acl_id)
		Int32ToBytes(buf[n:], int32(acl_id_len))
		n += 4
		copy(buf[n:], acl_id)
		n += acl_id_len
	}
	Int32ToBytes(buf[n:], req.flags)
	Int32ToBytes(buf[0:], int32(n))
	return int32(n + 4)
}

type createResponse struct {
	// 这里还有节点的路径，但目前不关心
}

func decodeCreateResponse(buf []byte, res *createResponse) {
}

func (zkCli *ZkCli) create(path string, data []byte) error {
	buf := make([]byte, BufferSize)
	n := encodeCreateRequest(buf, &createRequest{
		xid:    zkCli.getNextXid(),
		opcode: opCreate,
		path:   path,
		data:   data,
		acl:    WorldACL, // 默认
		flags:  0,        // 默认
	})
	req := &request{
		xid:    zkCli.xid,
		opcode: opCreate,
		reqbuf: buf[:n],
		resbuf: nil,
		err:    nil,
		done:   make(chan bool, 1),
	}
	zkCli.reqLock.Lock()
	zkCli.reqMap[req.xid] = req
	zkCli.reqLock.Unlock()
	zkCli.sentchan <- req
	<-req.done
	if req.err == nil {
		//res := &createResponse{}
		//decodeCreateResponse(req.res, res)
		//logger.Println(req.resbuf)
		return nil
	}
	// 这里应该判断各种错误
	return req.err
}

// API：新建节点
func (zk *ZkCli) Create(path string, data []byte) error {
	return zk.create(path, data)
}
