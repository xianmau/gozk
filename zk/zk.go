package zk

import (
	"net"
	"time"
)

const (
	DefaultPort    = 2181
	RecvTimeout    = 1                // 接收消息超时，单位：秒
	SessionTimeout = 4000             // 客户端会话超时，单位：毫秒
	PingInterval   = 2000             // Ping超时,单位：毫秒
	BufferSize     = 1024 * 1024 * 10 // 10M
	SentChanSize   = 64               //
	RecvChanSize   = 64               //
)

type ZkCli struct {
	protocolversion int32
	sessiontimeout  int32
	sessionid       int64
	password        []byte
	conn            *net.TCPConn
	state           int32
	sentchan        chan *request
}

func (zkCli *ZkCli) sentLoop() {
	// 设置心跳定时器
	pingTicker := time.NewTicker(PingInterval * time.Millisecond)
	defer pingTicker.Stop()

	buf := make([]byte, BufferSize)
	for {
		// 用select语句可以保证在高并发下心跳与请求不同时执行
		select {
		case req := <-zkCli.sentchan: // 捕获到客户端请求
			// 发送请求
			zkCli.conn.SetWriteDeadline(time.Now().Add(RecvTimeout * time.Second))
			_, err := zkCli.conn.Write(req.req)
			zkCli.conn.SetWriteDeadline(time.Time{})
			if err != nil {
				req.err = err
				req.done <- true
				continue
			}

			// 如果收到关闭连接的请求，则断开连接不再接收请求，暂不支持自动重连
			if req.opcode == opClose {
				req.err = nil
				req.done <- true
				return
			}

			// 获取响应
			zkCli.conn.SetReadDeadline(time.Now().Add(RecvTimeout * time.Second))
			_, err = zkCli.conn.Read(buf)
			zkCli.conn.SetReadDeadline(time.Time{})
			if err != nil {
				req.err = err
				req.done <- true
				continue
			}
			req.res = buf
			req.err = err
			req.done <- true
		case <-pingTicker.C: // 发送心跳
			zkCli.ping()
		}
	}
}

func (zkCli *ZkCli) connect(serverAddr string) error {
	// TCP拔号
	tcpAddr, err := net.ResolveTCPAddr("tcp4", serverAddr)
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return err // 连接超时
	}
	zkCli.conn = conn
	// 连接认证
	buf := make([]byte, 128)
	n := encodeConnectRequest(buf, &connectRequest{
		protocolversion: 0,
		lastzxidseen:    0,
		timeout:         0,
		sessionid:       0,
		password:        zkCli.password,
	})
	zkCli.conn.SetWriteDeadline(time.Now().Add(RecvTimeout * time.Second))
	_, err = zkCli.conn.Write(buf[:n])
	zkCli.conn.SetWriteDeadline(time.Time{})
	if err != nil {
		return err
	}
	zkCli.conn.SetReadDeadline(time.Now().Add(RecvTimeout * time.Second))
	_, err = zkCli.conn.Read(buf)
	zkCli.conn.SetReadDeadline(time.Time{})
	if err != nil {
		return err
	}
	res := &connectResponse{}
	decodeConnectResponse(buf[4:], res)
	zkCli.protocolversion = res.protocolversion
	zkCli.sessiontimeout = res.timeout
	zkCli.sessionid = res.sessionid
	zkCli.password = res.password
	// 循环发送请求
	go func() {
		zkCli.sentLoop()
	}()
	// 连接成功
	return nil
}

func (zkCli *ZkCli) close() error {
	buf := make([]byte, 128)
	n := encodeCloseRequest(buf, &closeRequest{
		xid:    0,
		opcode: opClose,
	})
	req := &request{
		xid:    0,
		opcode: opClose,
		req:    buf[:n],
		res:    nil,
		err:    nil,
		done:   make(chan bool, 1),
	}
	zkCli.sentchan <- req
	select {
	case <-req.done:
		if req.err == nil {
			close(zkCli.sentchan)
			zkCli.conn.Close()
			return nil
		}
		return req.err
	}
	return nil
}

func (zkCli *ZkCli) ping() error {
	buf := make([]byte, 128)
	n := encodePingRequest(buf, &pingRequest{
		xid:    -2,
		opcode: opPing,
	})
	zkCli.conn.SetWriteDeadline(time.Now().Add(RecvTimeout * time.Second))
	_, err := zkCli.conn.Write(buf[:n])
	zkCli.conn.SetWriteDeadline(time.Time{})
	if err != nil {
		logger.Println(err)
		return err
	}
	zkCli.conn.SetReadDeadline(time.Now().Add(RecvTimeout * time.Second))
	_, err = zkCli.conn.Read(buf)
	zkCli.conn.SetReadDeadline(time.Time{})
	if err != nil {
		logger.Println(err)
		return err
	}
	//res := &pingResponse{}
	//decodePingResponse(buf, res)
	return nil
}

func (zkCli *ZkCli) exists(path string) (bool, error) {
	buf := make([]byte, BufferSize)
	n := encodeExistRequest(buf, &existRequest{
		xid:    0,
		opcode: opExists,
		path:   path,
		watch:  false,
	})
	req := &request{
		xid:    0,
		opcode: opExists,
		req:    buf[:n],
		res:    nil,
		err:    nil,
		done:   make(chan bool, 1),
	}
	zkCli.sentchan <- req
	select {
	case <-req.done:
		if req.err == nil {
			buf = req.res
			res := &existResponse{}
			decodeExistResponse(buf, res)
			if res.errcode == 0 {
				return true, nil
			} else if res.errcode == 101 {
				return false, nil
			}
		}
		return false, req.err
	}
	return false, nil
}

func (zkCli *ZkCli) get(path string) ([]byte, error) {
	buf := make([]byte, BufferSize)
	n := encodeGetRequest(buf, &getRequest{
		xid:    0,
		opcode: opGet,
		path:   path,
		watch:  false,
	})
	req := &request{
		xid:    0,
		opcode: opGet,
		req:    buf[:n],
		res:    nil,
		err:    nil,
		done:   make(chan bool, 1),
	}
	zkCli.sentchan <- req
	select {
	case <-req.done:
		if req.err == nil {
			buf = req.res
			res := &getResponse{}
			decodeGetResponse(buf, res)
			return res.data, nil
		}
		return nil, req.err
	}
	return []byte{}, nil
}

func (zkCli *ZkCli) children(path string) ([]string, error) {
	buf := make([]byte, BufferSize)
	n := encodeChildrenRequest(buf, &childrenRequest{
		xid:    0,
		opcode: opChildren,
		path:   path,
		watch:  false,
	})
	req := &request{
		xid:    0,
		opcode: opChildren,
		req:    buf[:n],
		res:    nil,
		err:    nil,
		done:   make(chan bool, 1),
	}
	zkCli.sentchan <- req
	select {
	case <-req.done:
		if req.err == nil {
			buf = req.res
			res := &childrenResponse{}
			decodeChildrenResponse(buf, res)
			return res.children, nil
		}
		return nil, req.err
	}
	return []string{}, nil
}

func (zkCli *ZkCli) create(path string, data []byte) error {
	buf := make([]byte, BufferSize)
	n := encodeCreateRequest(buf, &createRequest{
		xid:    0,
		opcode: opCreate,
		path:   path,
		data:   data,
		acl:    WorldACL, // 默认
		flags:  0,        // 默认
	})
	req := &request{
		xid:    0,
		opcode: opCreate,
		req:    buf[:n],
		res:    nil,
		err:    nil,
		done:   make(chan bool, 1),
	}
	zkCli.sentchan <- req
	select {
	case <-req.done:
		//if req.err == nil {
		//	buf = req.res
		//	res := &createResponse{}
		//	decodeCreateResponse(buf, res)
		//	return nil
		//}
		return req.err
	}
	return nil
}

func (zkCli *ZkCli) set(path string, data []byte) error {
	buf := make([]byte, BufferSize)
	n := encodeSetRequest(buf, &setRequest{
		xid:     0,
		opcode:  opSet,
		path:    path,
		data:    data,
		version: -1, // 默认
	})
	req := &request{
		xid:    0,
		opcode: opSet,
		req:    buf[:n],
		res:    nil,
		err:    nil,
		done:   make(chan bool, 1),
	}
	zkCli.sentchan <- req
	select {
	case <-req.done:
		//if req.err == nil {
		//	buf = req.res
		//	res := &setResponse{}
		//	decodeSetResponse(buf, res)
		//	return nil
		//}
		return req.err
	}
	return nil
}

func (zkCli *ZkCli) delete(path string) error {
	buf := make([]byte, BufferSize)
	n := encodeDeleteRequest(buf, &deleteRequest{
		xid:     0,
		opcode:  opDelete,
		path:    path,
		version: -1, // 默认
	})
	req := &request{
		xid:    0,
		opcode: opDelete,
		req:    buf[:n],
		res:    nil,
		err:    nil,
		done:   make(chan bool, 1),
	}
	zkCli.sentchan <- req
	select {
	case <-req.done:
		//if req.err == nil {
		//	buf = req.res
		//	res := &deleteResponse{}
		//	decodeDeleteResponse(buf, res)
		//	return nil
		//}
		return req.err
	}
	return nil
}
