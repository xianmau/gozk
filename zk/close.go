package zk

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

type closeResponse struct {
}

func decodeCloseResponse(buf []byte, res *closeResponse) {
}

func (zkCli *ZkCli) close() error {
	buf := make([]byte, 12)
	n := encodeCloseRequest(buf, &closeRequest{
		xid:    zkCli.getNextXid(),
		opcode: opClose,
	})
	req := &request{
		xid:    zkCli.xid,
		opcode: opClose,
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
		close(zkCli.sentchan)
		zkCli.conn.Close()
		//logger.Println(req.resbuf)
		return nil
	}
	// 这里应该判断各种错误
	return req.err
}

// API：关闭连接
func (zk *ZkCli) Close() error {
	return zk.close()
}
