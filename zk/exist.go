package zk

type existRequest struct {
	xid    int32
	opcode int32
	path   string
	watch  bool
}

func encodeExistRequest(buf []byte, req *existRequest) int32 {
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

type existResponse struct {
	// 这里还有节点的状态，但目前不关心
}

func decodeExistResponse(buf []byte, res *existResponse) {
}

func (zkCli *ZkCli) exists(path string) (bool, error) {
	buf := make([]byte, BufferSize)
	n := encodeExistRequest(buf, &existRequest{
		xid:    zkCli.getNextXid(),
		opcode: opExists,
		path:   path,
		watch:  false,
	})
	req := &request{
		xid:    zkCli.xid,
		opcode: opExists,
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
		if req.resheader.errcode == 0 {
			return true, nil
		} else if req.resheader.errcode == 101 {
			return false, nil
		}
	}
	return false, req.err
}

// API：测试节点是否存在
func (zk *ZkCli) Exists(path string) (bool, error) {
	return zk.exists(path)
}
