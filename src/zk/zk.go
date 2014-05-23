package zk

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// API:Connect
func Connect(servers []string, recvTimeout time.Duration) (*ZK, error) {
	// 处理没有带端口号的地址
	for _, serverip := range servers {
		if strings.ContainsRune(serverip, rune(':')) {
			serverip = fmt.Sprintf("%s:%d", serverip, defaultPort)
		}
	}
	// 初始化一个实例
	zk := ZK{
		servers:        servers,
		serversIndex:   0,
		conn:           nil,
		connectTimeout: 1 * time.Second,
		sessionId:      0,
		sessionTimeout: 86400, // 一天，呵呵
		password:       emptyPassword,
		state:          StateDisconnected,

		pingInterval: time.Duration((int64(recvTimeout) >> 1)),
		recvTimeout:  recvTimeout,

		sendChan: make(chan *request, sendChanSize),
		requests: make(map[int32]*request),
	}
	// 开个协程来连接
	var err error = nil
	go func() {
		err = zk.connect(servers, recvTimeout)
	}()
	// 这是其实应该是有问题的，因为是异步连接的，err早就返回了，根本接收不到什么错误
	return &zk, err
}

// 进行TCP拔号
func (zk *ZK) dial() error {
	var conn net.Conn
	var err error = nil
	zk.state = StateConnecting
	// 尝试所有IP，一有成功拔号的，马上跳出
	for index, serverip := range servers {
		conn, err = net.DialTimeout("tcp", serverip, zk.connectTimeout)
		if err == nil {
			zk.conn = conn
			zk.serversIndex = index
			zk.state = StateConnected
			return nil
		}
		log.Printf("Failed to connect to %s: %+v", c.servers[c.serverIndex], err)
	}
	// 报告所有IP不通
	return ErrZoneDown
}

// 进行连接认证
func (zk *ZK) authenticate() error {
	buf := make([]byte, 256)

	connectRequest := &connectRequest{
		ProtocolVersion: protocolVersion,
		LastZxidSeen:    zk.lastZxid,
		TimeOut:         zk.sessionTimeout,
		SessionId:       zk.sessionId,
		Passwd:          zk.password,
	}

	var n int
	n, err = encodePacket(buf[4:], connectRequest)
	binary.BigEndian.PutUint32(buf[:4], uint32(n))
	_, err = zk.conn.Write(buf[:n+4])
	if err != nil {
		return err
	}
	_, err = io.ReadFull(zk.conn, buf[:4])
	if err != nil {
		return err
	}
	n = int(binary.BigEndian.Uint32(buf[:4]))
	if cap(buf) < n {
		buf = make([]byte, n)
	}
	_, err = io.ReadFull(zk.conn, buf[:n])
	if err != nil {
		return err
	}

	r := connectResponse{}
	_, err = decodePacket(buf[:n], &r)
	if err != nil {
		return err
	}
	// 超时
	if r.SessionId == 0 {
		zk.sessionId = 0
		zk.password = emptyPassword
		zk.lastZxid = 0
		return ErrSessionExpired
	}
	// 如果是新的会话，那么就设置xid为0
	if zk.sessionId != r.SessionId {
		atomic.StoreInt32(&zk.xid, 0)
	}
	zk.sessionTimeout = r.TimeOut
	zk.sessionId = r.SessionId
	zk.password = r.Passwd
	zk.state = StateHasSession

	return nil
}

// 连接到ZK服务器，如果成功，返回一个ZK实例
func (zk *ZK) connect(servers []string, recvTimeout time.Duration) error {
	// 先来个无限循环，这样某台机宕掉了，程序就会尝试其它机
	// 不过这样的话，如果所有机都宕了（所有IP不通），那不就是死循环了？！
	// 所以先这样搞，就是将主机列表都扫过一遍后如果还没成，就退出算了
	for {
		// 如果所有IP都没拔通，则报错退出
		err := zk.dial()
		if err != nil {
			return err
		}
		// 进行认证
		err = zk.authenticate()

		switch {
		case err != nil && zk.conn != nil: // 说明拔号成功，但是认证失败，关闭TCP连接
			zk.conn.Close()
		case err == nil: // 说明认证成功
			closeChan := make(chan bool)

			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				zk.sendLoop(zk.conn, closeChan)
				zk.conn.Close()
				wg.Done()
			}()
			wg.Add(1)
			go func() {
				err = zk.recvLoop(zk.conn)
				if err == nil {
					panic("zk: recvLoop should never return nil error")
				}
				close(closeChan)
				wg.Done()
			}()
			wg.Wait()
		}

		zk.state = StateDisconnected

		if err != io.EOF && err != ErrSessionExpired && !strings.Contains(err.Error(), "use of closed network connection") {
			log.Println(err)
		}

		select {
		case <-zk.shouldQuit:
			zk.flushRequests(ErrClosing)
			return ErrClosing
		}
		// 接下来应该是监听是不是退出了，然后准备重连
	}
	// 如果跑到这来说明什么呢
	return err
}

// 获取节点数据
func (zk *ZK) get(path string) ([]byte, *Stat, error) {
	res := &getDataResponse{}

	_, err := zk.request(opGetData, &getDataRequest{Path: path, Watch: false}, res, nil)
	return res.Data, &res.Stat, err
}

// 新建一个节点
func (zk *ZK) create(path string, data []byte) {
	// TODO:
}

// 断开与服务器的连接
func (zk *ZK) close() {
	// TODO:
}

func (zk *ZK) queueRequest(opcode int32, req interface{}, res interface{}, recvFunc func(*request, *responseHeader, error)) <-chan response {
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
	r := <-zk.queueRequest(opcode, req, res, recvFunc)
	return r.zxid, r.err
}

// 感觉这个就是来对应消息队列里的顺序
func (zk *ZK) nextXid() int32 {
	return atomic.AddInt32(&zk.xid, 1)
}
