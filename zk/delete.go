package zk

type deleteRequest struct {
	xid     int32
	opcode  int32
	path    string
	version int32
}

func encodeDeleteRequest(buf []byte, req *deleteRequest) int32 {
	path := []byte(req.path)
	path_len := int32(len(path))
	n := 4
	Int32ToBytes(buf[n:], req.xid)
	n += 4
	Int32ToBytes(buf[n:], req.opcode)
	n += 4
	Int32ToBytes(buf[n:], path_len)
	n += 4
	copy(buf[n:], path)
	n += int(path_len)
	Int32ToBytes(buf[n:], req.version)
	Int32ToBytes(buf[0:], int32(n))
	return int32(n + 4)
}

type deleteResponse struct {
}

func decodeDeleteResponse(buf []byte, res *deleteResponse) {
}

func (zkCli *ZkCli) delete(path string) error {
	buf := make([]byte, BufferSize)
	n := encodeDeleteRequest(buf, &deleteRequest{
		xid:     zkCli.getNextXid(),
		opcode:  opDelete,
		path:    path,
		version: -1, // 默认
	})
	req := &request{
		xid:    zkCli.xid,
		opcode: opDelete,
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
		//res := &deleteResponse{}
		//decodeDeleteResponse(req.res, res)
		logger.Println(req.resbuf)
		return nil
	}
	// 这里应该判断各种错误
	return req.err
}

// API：删除节点
func (zk *ZkCli) Delete(path string) error {
	return zk.delete(path)
}

// API：递归删除节点
func (zk *ZkCli) DeleteRecur(path string) error {
	if flag, err := zk.Exists(path); err == nil && !flag {
		return err
	}
	children, err := zk.Children(path)
	if err != nil {
		return err
	}
	for _, znode := range children {
		sub_znode := path + "/" + znode
		zk.DeleteRecur(sub_znode)
	}
	return zk.delete(path)
}
