package zk

type childrenRequest struct {
	xid    int32
	opcode int32
	path   string
	watch  bool
}

func encodeChildrenRequest(buf []byte, req *childrenRequest) int32 {
	path := []byte(req.path)
	path_len := int32(len(path))
	Int32ToBytes(buf[4:], req.xid)
	Int32ToBytes(buf[8:], req.opcode)
	Int32ToBytes(buf[12:], path_len)
	copy(buf[16:], req.path)
	BoolToBytes(buf[16+path_len:], req.watch)
	Int32ToBytes(buf[0:], 13+path_len)
	return 17 + path_len
}

type childrenResponse struct {
	children []string
}

func decodeChildrenResponse(buf []byte, res *childrenResponse) {
	children_cnt := BytesToInt32(buf)
	for n := 4; children_cnt > 0; children_cnt-- {
		child_len := int(BytesToInt32(buf[n:]))
		n += 4
		res.children = append(res.children, string(buf[n:n+child_len]))
		n += child_len
	}
}

func (zkCli *ZkCli) children(path string) ([]string, error) {
	buf := make([]byte, BufferSize)
	n := encodeChildrenRequest(buf, &childrenRequest{
		xid:    zkCli.getNextXid(),
		opcode: opChildren,
		path:   path,
		watch:  false,
	})
	req := &request{
		xid:    zkCli.xid,
		opcode: opChildren,
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
		res := &childrenResponse{
			children: []string{},
		}
		decodeChildrenResponse(req.resbuf, res)
		//logger.Println(req.resbuf)
		return res.children, nil
	}
	// 这里应该判断各种错误
	return nil, req.err
}

// API：获取子节点列表
func (zk *ZkCli) Children(path string) ([]string, error) {
	return zk.children(path)
}
