package zk

type getRequest struct {
	xid    int32
	opcode int32
	path   string
	watch  bool
}

func encodeGetRequest(buf []byte, req *getRequest) int32 {
	path := []byte(req.path)
	path_len := int32(len(path))
	Int32ToBytes(buf[4:], req.xid)
	Int32ToBytes(buf[8:], req.opcode)
	Int32ToBytes(buf[12:], path_len)
	copy(buf[16:], path)
	BoolToBytes(buf[16+path_len:], req.watch)
	Int32ToBytes(buf[0:], 13+path_len)
	return 17 + path_len
}

type getResponse struct {
	data []byte
	// 这里还有节点的状态，但目前不关心
}

func decodeGetResponse(buf []byte, res *getResponse) {
	data_len := BytesToInt32(buf)
	res.data = buf[4 : 4+data_len]
}

func (zkCli *ZkCli) get(path string) ([]byte, error) {
	buf := make([]byte, BufferSize)
	n := encodeGetRequest(buf, &getRequest{
		xid:    zkCli.getNextXid(),
		opcode: opGet,
		path:   path,
		watch:  false,
	})
	req := &request{
		xid:    zkCli.xid,
		opcode: opGet,
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
		res := &getResponse{}
		decodeGetResponse(req.resbuf, res)
		//logger.Println(req.resbuf)
		return res.data, nil
	}
	// 这里应该判断各种错误
	return nil, req.err
}

// API：获取节点数据
func (zk *ZkCli) Get(path string) ([]byte, error) {
	return zk.get(path)
}
