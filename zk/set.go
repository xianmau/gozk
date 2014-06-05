package zk

type setRequest struct {
	xid     int32
	opcode  int32
	path    string
	data    []byte
	version int32
}

func encodeSetRequest(buf []byte, req *setRequest) int32 {
	path := []byte(req.path)
	path_len := int32(len(path))
	data_len := int32(len(req.data))
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
	Int32ToBytes(buf[n:], req.version)
	Int32ToBytes(buf[0:], int32(n))
	return int32(n + 4)
}

type setResponse struct {
	// 这里还有节点的状态，但目前不关心
}

func decodeSetResponse(buf []byte, res *setResponse) {
}

func (zkCli *ZkCli) set(path string, data []byte) error {
	buf := make([]byte, BufferSize)
	n := encodeSetRequest(buf, &setRequest{
		xid:     zkCli.getNextXid(),
		opcode:  opSet,
		path:    path,
		data:    data,
		version: -1, // 默认
	})
	req := &request{
		xid:    zkCli.xid,
		opcode: opSet,
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
		//res := &setResponse{}
		//decodeSetResponse(req.res, res)
		//logger.Println(req.resbuf)
		return nil
	}
	return req.err
}

// API：设置节点数据
func (zk *ZkCli) Set(path string, data []byte) error {
	return zk.set(path, data)
}
