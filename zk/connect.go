package zk

import (
	"fmt"
	"io"
	"net"
	"strings"
	"time"
)

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

type connectResponse struct {
	protocolversion int32
	timeout         int32
	sessionid       int64
	password        []byte
}

func decodeConnectResponse(buf []byte, res *connectResponse) {
	res.protocolversion = BytesToInt32(buf)
	res.timeout = BytesToInt32(buf[4:])
	res.sessionid = BytesToInt64(buf[8:])
	n := BytesToInt32(buf[16:])
	res.password = buf[20 : 20+n]
}

func (zkCli *ZkCli) sentLoop() error {
	// 设置心跳定时器
	pingTicker := time.NewTicker(PingInterval * time.Millisecond)
	defer pingTicker.Stop()
	pingBuf := make([]byte, 12)
	for {
		select {
		case req := <-zkCli.sentchan: // 收到客户端请求
			zkCli.conn.SetWriteDeadline(time.Now().Add(RecvTimeout * time.Second))
			_, err := zkCli.conn.Write(req.reqbuf)
			if err != nil {
				logger.Println(err)
				return err
			}
			zkCli.conn.SetWriteDeadline(time.Time{})

			if req.opcode == opClose {
				req.err = nil
				req.done <- true
				logger.Println(zkCli.reqMap)
				return nil
			}
		case <-pingTicker.C: // 发送心跳
			encodePingRequest(pingBuf, &pingRequest{
				xid:    -2,
				opcode: opPing,
			})
			zkCli.conn.SetWriteDeadline(time.Now().Add(RecvTimeout * time.Second))
			_, err := zkCli.conn.Write(pingBuf)
			zkCli.conn.SetWriteDeadline(time.Time{})
			if err != nil {
				logger.Println(err)
				return err
			}
		}
	}
}

func (zkCli *ZkCli) recvLoop() {
	pkgSizeBuf := make([]byte, 4)
	for {
		zkCli.conn.SetReadDeadline(time.Now().Add(RecvTimeout * time.Second))
		//_, err := zkCli.conn.Read(pkgSizeBuf)
		_, err := io.ReadFull(zkCli.conn, pkgSizeBuf)
		zkCli.conn.SetReadDeadline(time.Time{})
		if err != nil {
			logger.Println(err)
			return
		}
		pkgSize := BytesToInt32(pkgSizeBuf)
		pkgBuf := make([]byte, pkgSize)
		//_, err = zkCli.conn.Read(pkgBuf)
		_, err = io.ReadFull(zkCli.conn, pkgBuf)
		if err != nil {
			logger.Println(err)
			return
		}

		resHeader := &responseHeader{}
		decodeResponseHeader(pkgBuf[:16], resHeader)

		if resHeader.xid == -2 {
			// ping pkg
		} else if resHeader.xid > 0 {
			zkCli.reqLock.Lock()
			if req, ok := zkCli.reqMap[resHeader.xid]; ok {
				req.resbuf = pkgBuf[16:pkgSize]
				req.resheader = resHeader
				req.done <- true
				delete(zkCli.reqMap, resHeader.xid)
			}
			zkCli.reqLock.Unlock()
		}
	}
}

func (zkCli *ZkCli) connect(serverAddr string) error {
	// TCP拔号，超时大概在20s
	tcpAddr, err := net.ResolveTCPAddr("tcp4", serverAddr)
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return err // 连接超时
	}
	zkCli.conn = conn

	// 连接认证
	buf := make([]byte, 48)
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
	res := &connectResponse{}
	decodeConnectResponse(buf[4:], res)
	zkCli.protocolversion = res.protocolversion
	zkCli.sessiontimeout = res.timeout
	zkCli.sessionid = res.sessionid
	zkCli.password = res.password
	zkCli.state = stateConnected

	// 发送请求
	go func() {
		err := zkCli.sentLoop()
		if err == nil {
			// 正常关闭连接成功
		} else {
			// 其它原因
		}
	}()
	// 接收响应
	go func() {
		zkCli.recvLoop()
		logger.Println("recvloop exit")
	}()

	return nil
}

// API：连接
func (zk *ZkCli) Connect(servers []string) error {
	for i, serverip := range servers {
		if !strings.ContainsRune(serverip, rune(':')) {
			servers[i] = fmt.Sprintf("%s:%d", serverip, DefaultPort)
		}
	}
	zk.state = stateConnecting
	for i := range servers {
		err := zk.connect(servers[i])
		if err == nil {
			zk.state = stateConnected
			return nil
		}
		//return err
	}
	zk.state = stateDisconnect
	return errMap[errConnectionDisabled]
}
