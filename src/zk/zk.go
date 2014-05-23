package zk

import (
	"fmt"
	"net"
	//"strconv"
	//"strings"
	"encoding/binary"
	"io"
	"sync/atomic"
	"time"
)

// 连接到服务器
func Connect(servers []string, connectTimeout time.Duration) (ZK, error) {
	// 初始化一个ZK实例
	zk := ZK{
		servers:        servers,
		connectTimeout: connectTimeout,
		sessionId:      0,
		sessionTimeout: 30000,
		conn:           nil,
		password:       emptyPassword,
	}

	var conn net.Conn
	var err error = nil

	// 尝试所有IP，一有成功连接的，马上跳出
	for _, serverip := range servers {
		conn, err = net.DialTimeout("tcp", serverip, connectTimeout)
		if err == nil {
			zk.conn = conn
			break
		}
	}

	// 如果连接成功，则进行认证
	if zk.conn != nil {
		buf := make([]byte, 256)
		// 新建连接认证请求
		zk.connectRequest = &connectRequest{
			ProtocolVersion: protocolVersion,
			LastZxidSeen:    zk.lastZxid,
			TimeOut:         zk.sessionTimeout,
			SessionId:       zk.sessionId,
			Passwd:          zk.password,
		}
		var n int
		// 进行请求编码
		n, err = encodePacket(buf[4:], zk.connectRequest)
		binary.BigEndian.PutUint32(buf[:4], uint32(n))
		// 发送数据到服务器
		_, err = zk.conn.Write(buf[:n+4])
		if err != nil {
			return zk, err
		}
		// 读取数据长度
		_, err = io.ReadFull(zk.conn, buf[:4])
		if err != nil {
			return zk, err
		}
		n = int(binary.BigEndian.Uint32(buf[:4]))
		// 如果缓冲区太小，则扩容
		if cap(buf) < n {
			buf = make([]byte, n)
		}
		// 继续读取数据
		_, err = io.ReadFull(zk.conn, buf[:n])
		if err != nil {
			return zk, err
		}
		// 对读到的数据进行解码
		r := connectResponse{}
		_, err = decodePacket(buf[:n], &r)
		if err != nil {
			return zk, err
		}
		if r.SessionId == 0 {
			zk.sessionId = 0
			zk.password = emptyPassword
			zk.lastZxid = 0
			return zk, ErrSessionExpired
		}
		if zk.sessionId != r.SessionId {
			atomic.StoreInt32(&zk.xid, 0)
		}
		zk.sessionTimeout = r.TimeOut
		zk.sessionId = r.SessionId
		zk.password = r.Passwd

		fmt.Printf("%+v\n", zk)
	}
	return zk, err
}

// 获取节点数据
func (zk *ZK) Get(path string) ([]byte, *Stat, error) {
	res := &getDataResponse{}

	//_, err := zk.request(opGetData, &getDataRequest{Path: path, Watch: false}, res, nil)
	return res.Data, &res.Stat, err
}

// 新建一个节点
func (zk *ZK) Create(path string, data []byte) {

}

func (zk *Conn) queueRequest(opcode int32, req interface{}, res interface{}, recvFunc func(*request, *responseHeader, error)) <-chan response {
	rq := &request{
		xid:        zk.nextXid(),
		opcode:     opcode,
		pkt:        req,
		recvStruct: res,
		recvChan:   make(chan response, 1),
		recvFunc:   recvFunc,
	}
	zk.sendChan <- rq
	return rq.recvChan
}

func (zk *ZK) request(opcode int32, req interface{}, res interface{}, recvFunc func(*request, *responseHeader, error)) (int64, error) {
	r := <-zk.queryRequest(opcode, req, res, recvFunc)
	return r.zxid, r.err
}

// 断开与服务器的连接
func (zk *ZK) Close() {
	// TODO:
}
