package zk

import (
	"net"
	"time"
)

const (
	DefaultPort    = 2181
	RecvTimeout    = 1          // 接收消息超时，单位：秒
	SessionTimeout = 40000      // 客户端会话超时，单位：秒
	BufferSize     = 512 * 1024 // 512KB
)

type ZkCli struct {
	protocolversion int32
	sessiontimeout  int32
	sessionid       int64
	password        []byte
	conn            *net.TCPConn
	state           int32
}

func (zkCli *ZkCli) connect(serverAddr string) error {
	// TCP拔号
	tcpAddr, err := net.ResolveTCPAddr("tcp4", serverAddr)
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return err // 算连接超时
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
	return nil
}

func (zkCli *ZkCli) close() error {
	buf := make([]byte, 128)
	n := encodeCloseRequest(buf, &closeRequest{
		xid:    0,
		opcode: opClose,
	})
	zkCli.conn.SetWriteDeadline(time.Now().Add(RecvTimeout * time.Second))
	_, err := zkCli.conn.Write(buf[:n])
	zkCli.conn.SetWriteDeadline(time.Time{})
	if err != nil {
		logger.Println(err)
		return err
	}
	err = zkCli.conn.Close()
	if err != nil {
		logger.Println(err)
		return err
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
	res := &pingResponse{}
	decodePingResponse(buf[4:], res)
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
	zkCli.conn.SetWriteDeadline(time.Now().Add(RecvTimeout * time.Second))
	_, err := zkCli.conn.Write(buf[:n])
	zkCli.conn.SetWriteDeadline(time.Time{})
	if err != nil {
		logger.Println(err)
		return false, err
	}
	zkCli.conn.SetReadDeadline(time.Now().Add(RecvTimeout * time.Second))
	_, err = zkCli.conn.Read(buf)
	zkCli.conn.SetReadDeadline(time.Time{})
	if err != nil {
		logger.Println(err)
		return false, err
	}
	res := &existResponse{}
	decodeExistResponse(buf[4:], res)
	if res.errcode == 0 {
		return true, nil
	} else if res.errcode == 101 {
		return false, nil
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
	zkCli.conn.SetWriteDeadline(time.Now().Add(RecvTimeout * time.Second))
	_, err := zkCli.conn.Write(buf[:n])
	zkCli.conn.SetWriteDeadline(time.Time{})
	if err != nil {
		logger.Println(err)
		return nil, err
	}
	zkCli.conn.SetReadDeadline(time.Now().Add(RecvTimeout * time.Second))
	_, err = zkCli.conn.Read(buf)
	zkCli.conn.SetReadDeadline(time.Time{})
	if err != nil {
		logger.Println(err)
		return nil, err
	}
	res := &getResponse{}
	decodeGetResponse(buf, res)
	return res.data, nil
}

func (zkCli *ZkCli) children(path string) ([]string, error) {
	buf := make([]byte, BufferSize)
	n := encodeChildrenRequest(buf, &childrenRequest{
		xid:    0,
		opcode: opChildren,
		path:   path,
		watch:  false,
	})
	zkCli.conn.SetWriteDeadline(time.Now().Add(RecvTimeout * time.Second))
	_, err := zkCli.conn.Write(buf[:n])
	zkCli.conn.SetWriteDeadline(time.Time{})
	if err != nil {
		logger.Println(err)
		return nil, err
	}
	zkCli.conn.SetReadDeadline(time.Now().Add(RecvTimeout * time.Second))
	_, err = zkCli.conn.Read(buf)
	zkCli.conn.SetReadDeadline(time.Time{})
	if err != nil {
		logger.Println(err)
		return nil, err
	}
	res := &childrenResponse{}
	decodeChildrenResponse(buf, res)
	return res.children, nil
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
	res := &createResponse{}
	decodeCreateResponse(buf[4:], res)
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
	res := &setResponse{}
	decodeSetResponse(buf[4:], res)
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
	res := &deleteResponse{}
	decodeDeleteResponse(buf[4:], res)
	return nil
}

/*----------------------------------------------
 *
 * 以下是各个协议的结构以及它们的编码和解码
 *
 ----------------------------------------------*/
