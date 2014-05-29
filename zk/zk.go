package zk

import (
	"encoding/binary"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Zookeeper的数据结构
type ZK struct {
	lastZxid          int64              // 最后修改的节点
	xid               int32              // 用来标记请求的版本号，是一个递增值
	servers           []string           // 服务器地址集合
	serversIndex      int                // 当前连接的服务器序号
	conn              net.Conn           // 标准库里的TCP连接结构体
	connectTimeout    time.Duration      // 连接超时，其实是指拔号超时
	sessionId         int64              // 会话Id
	sessionTimeout    int32              // 会话超时
	password          []byte             // 密码
	state             int32              // 连接状态
	heartbeatInterval time.Duration      // 心跳间隔，一般为接收超时的一半
	recvTimeout       time.Duration      // 接收超时，一般为心跳的两倍
	shouldQuit        chan bool          // 是否退出
	sendChan          chan *request      // 请求信道，客户端把要发送的请求丢到这里，可以实现同步
	requests          map[int32]*request // 正在准备的请求映射，在请求结构中有Xid，可用来对应哪个请求
	requestsLock      sync.Mutex         // 请求锁
	reconnectDelay    time.Duration      // 延迟重连时间

}

// Znode状态的数据结构，基本不想理它
type Stat struct {
	Czxid          int64 // 节点被创建的Zxid值
	Mzxid          int64 // 节点被修改的Zxid值
	Ctime          int64 // 节点被创建的时间
	Mtime          int64 // 节点被修改的时间
	Version        int32 // 节点数据修改的版本号
	CVersion       int32 // 节点的子节点被修改的版本号
	AVersion       int32 // 节点的ACL被修改的版本号
	EphemeralOwner int64 // 如果不是临时节点，值为0，否则为节点拥有者的会话Id
	DataLength     int32 // 节点数据长度
	NumChildren    int32 // 节点的子节点数量
	Pzxid          int64 // 最后修改子节点
}

// TCP拔号，只要有一个IP拔通就正常返回，否则报错
func (zk *ZK) dial() error {
	zk.state = StateConnecting
	for index, serverip := range zk.servers {
		conn, err := net.DialTimeout("tcp", serverip, zk.connectTimeout)
		if err == nil {
			zk.conn = conn
			zk.serversIndex = index
			zk.state = StateConnected
			return nil
		}
		log.Printf("%+v", err)
	}
	return ErrDialingFaild
}

// 连接认证
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
	n, err := encodePacket(buf[4:], connectRequest)
	binary.BigEndian.PutUint32(buf[:4], uint32(n))
	_, err = zk.conn.Write(buf[:n+4])
	if err != nil {
		return err // 写出错
	}
	_, err = io.ReadFull(zk.conn, buf[:4])
	if err != nil {
		return err // 读出错
	}
	n = int(binary.BigEndian.Uint32(buf[:4]))
	if cap(buf) < n {
		buf = make([]byte, n)
	}
	_, err = io.ReadFull(zk.conn, buf[:n])
	if err != nil {
		return err // 读出错
	}
	recv := connectResponse{}
	_, err = decodePacket(buf[:n], &recv)
	if err != nil {
		return err // 解码出错
	}
	if recv.SessionId == 0 {
		zk.sessionId = 0
		zk.password = emptyPassword
		zk.lastZxid = 0
		return ErrSessionExpired // 超时
	}

	atomic.StoreInt32(&zk.xid, 0)
	zk.sessionTimeout = recv.TimeOut
	zk.sessionId = recv.SessionId
	zk.password = recv.Passwd
	zk.state = StateHasSession

	return nil
}

// 发送错误信息到所有正在准备的请求，并清空请求映射
func (zk *ZK) flushRequests(err error) {
	zk.requestsLock.Lock()
	for _, req := range zk.requests {
		req.recvChan <- response{-1, err} // 设zxid为-1的意思是让它发生在任何其它zxid之前
	}
	zk.requests = make(map[int32]*request)
	zk.requestsLock.Unlock()
}

// 循环发送请求
func (zk *ZK) sendLoop(closeChan <-chan bool) error {
	// 设置一个心跳定时器
	heartbeetTicker := time.NewTicker(zk.heartbeatInterval)
	defer heartbeetTicker.Stop()

	buf := make([]byte, bufferSize)
	for {

		select {
		// 等待获取一个请求，然后进行处理
		case req := <-zk.sendChan:
			reqHeader := &requestHeader{req.xid, req.opcode}
			headerLen, err := encodePacket(buf[4:], reqHeader) // 将请求头进行编码，其实这里返回4，因为xid和opcode都是2字节的类型
			if err != nil {
				req.recvChan <- response{-1, err}
				continue
			}
			bodyLen, err := encodePacket(buf[4+headerLen:], req.packet) // 将请求头进行编码，这里就不知道多长了
			if err != nil {
				req.recvChan <- response{-1, err}
				continue
			}
			packetLen := headerLen + bodyLen
			binary.BigEndian.PutUint32(buf[:4], uint32(packetLen))

			// 如果收到关闭连接的请求，则接收端将收到连接关闭的错误
			zk.requestsLock.Lock()
			select {
			case <-closeChan:
				req.recvChan <- response{-1, ErrConnectionClosed}
				zk.requestsLock.Unlock()
				return ErrConnectionClosed
			default:
			}
			zk.requests[req.xid] = req
			zk.requestsLock.Unlock()

			zk.conn.SetWriteDeadline(time.Now().Add(zk.recvTimeout))
			_, err = zk.conn.Write(buf[:packetLen+4])
			zk.conn.SetWriteDeadline(time.Time{})
			if err != nil {
				req.recvChan <- response{-1, err}
				zk.conn.Close()
				return err
			}
		case <-heartbeetTicker.C:
			packetLen, err := encodePacket(buf[4:], &requestHeader{xid: -2, opcode: opPing})
			if err != nil {
				panic("zk: opPing should never fail to serialize")
			}
			binary.BigEndian.PutUint32(buf[:4], uint32(packetLen))
			zk.conn.SetWriteDeadline(time.Now().Add(zk.recvTimeout))
			_, err = zk.conn.Write(buf[:packetLen+4])
			zk.conn.SetWriteDeadline(time.Time{})
			if err != nil {
				zk.conn.Close()
				return err
			}
		case <-closeChan: // 如果等到closeChan有true的值传来，那么就正常结束
			return nil
		}
	}
}

// 循环接收服务器消息
func (zk *ZK) recvLoop() error {
	buf := make([]byte, bufferSize)
	for {
		zk.conn.SetReadDeadline(time.Now().Add(zk.recvTimeout))
		_, err := io.ReadFull(zk.conn, buf[:4])
		if err != nil {
			return err
		}
		packetLen := int(binary.BigEndian.Uint32(buf[:4]))
		if cap(buf) < packetLen {
			buf = make([]byte, packetLen)
		}
		_, err = io.ReadFull(zk.conn, buf[:packetLen])
		if err != nil {
			return err
		}
		zk.conn.SetReadDeadline(time.Time{})

		res := responseHeader{}
		_, err = decodePacket(buf[:16], &res) // Xid占4字节，Zxid占8字节，Err占4字节
		if err != nil {
			return err
		}

		// 根据Xid进行相应的操作
		if res.Xid == -1 { // 表示watch事件，暂不处理

		} else if res.Xid == -2 { // 表示ping响应，直接忽略

		} else if res.Xid < 0 {
			log.Printf("Xid < 0 (%d) but not ping or watcher event", res.Xid)
		} else {
			if res.Xid > 0 {
				zk.lastZxid = res.Zxid // 更新一下zxid
			}

			zk.requestsLock.Lock()
			req, ok := zk.requests[res.Xid]
			if ok {
				delete(zk.requests, res.Xid) // 删除已经响应的请求
			}
			zk.requestsLock.Unlock()

			if !ok {
				log.Printf("Response for unknown request with xid %d", res.Xid)
			} else {
				if res.Err != 0 {
					err = errCodeToError[res.Err]
				} else {
					_, err = decodePacket(buf[16:16+packetLen], req.recvStruct)
				}
				if req.recvFunc != nil { // 回调方法，暂时用不上
					req.recvFunc(req, &res, err)
				}
				req.recvChan <- response{res.Zxid, err}
				if req.opcode == opClose {
					return io.EOF
				}
			}
		}
	}
	return nil
}

// 请求队列
func (zk *ZK) queueRequest(opcode int32, req interface{}, res interface{}, recvFunc func(*request, *responseHeader, error)) <-chan response {
	rq := &request{
		xid:        zk.nextXid(),
		opcode:     opcode,
		packet:     req,
		recvStruct: res,
		recvChan:   make(chan response, 1), // 信道大小为1，也就保证一个请求最多一个响应
		recvFunc:   recvFunc,
	}
	zk.sendChan <- rq
	return rq.recvChan
}

// 当一个请求来时，直接将它丢进请求队列中
func (zk *ZK) request(opcode int32, req interface{}, res interface{}, recvFunc func(*request, *responseHeader, error)) (int64, error) {
	r := <-zk.queueRequest(opcode, req, res, recvFunc)
	return r.zxid, r.err
}

// 类似自增Id，用来对应请求和响应
func (zk *ZK) nextXid() int32 {
	return atomic.AddInt32(&zk.xid, 1)
}

// 连接到ZK服务器，这是个无限循环，如果当前机子宕掉，程序就会尝试其它的机子
func (zk *ZK) connect(servers []string, recvTimeout time.Duration) {
	for {
		// 进行拔号，若失败，则按设置的延迟时间进行重试
		err := zk.dial()
		if err == ErrDialingFaild {
			log.Printf("%+v", err)
			time.Sleep(zk.reconnectDelay)
			continue
		}
		// 拔通则进行认证
		err = zk.authenticate()
		// 判断是什么错误
		switch {
		case err != nil && zk.conn != nil: // 拔号成功，认证失败，关闭TCP连接，准备重试
			zk.conn.Close()
		case err == nil: // 认证成功
			// 创建一个关闭信道，然后丢到sendLoop里，在那里通过select等待关闭
			closeChan := make(chan bool)
			// 采用锁来同步，WaitGroup，就是等待执行完一组方法后，才继续执行
			// 大概流程就是，客户端至少不断发送心跳包到服务器，然后还有什么查看啊添加啊删除啊什么的
			// 同时，客户端从服务器那边获取响应，然后处理
			// 这个过程按理说是无限进行的，因为至少有心跳
			// 如果出现超时什么的就会停止，跑到Wait()后面继续执行
			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				zk.sendLoop(closeChan)
				zk.conn.Close()
				wg.Done()
			}()
			wg.Add(1)
			go func() {
				err = zk.recvLoop()
				if err == nil {
					panic("zk: recvLoop should never return nil error")
				}
				close(closeChan)
				wg.Done()
			}()
			wg.Wait()
		}

		zk.state = StateDisconnected // 来到这里说明要不认证失败，要不超时了

		// 如果不是文件结束或者不是会话超时或者连接关闭，就记录一条日志
		if err != io.EOF && err != ErrSessionExpired && !strings.Contains(err.Error(), "use of closed network connection") {
			log.Println(err)
		}

		select {
		case <-zk.shouldQuit:
			zk.flushRequests(ErrClosing)
			return
		default: // 这句是让select变成非阻塞的
		}
	}
}

// 断开与服务器的连接
func (zk *ZK) close() {
	close(zk.shouldQuit) // 关闭退出信道
	select {
	case <-zk.queueRequest(opClose, &closeRequest{}, &closeResponse{}, nil):
	case <-time.After(time.Second):
	}
}

// 判断节点是否存在
func (zk *ZK) exists(path string) (bool, *Stat, error) {
	res := &existsResponse{}
	_, err := zk.request(opExists, &existsRequest{Path: path, Watch: false}, res, nil)
	exists := true
	if err == ErrNoNode {
		exists = false
		err = nil
	}
	return exists, &res.Stat, err
}

// 获取节点数据
func (zk *ZK) get(path string) ([]byte, *Stat, error) {
	res := &getDataResponse{}
	_, err := zk.request(opGetData, &getDataRequest{Path: path, Watch: false}, res, nil)
	return res.Data, &res.Stat, err
}

// 设置节点数据
func (zk *ZK) set(path string, data []byte, version int32) (*Stat, error) {
	res := &setDataResponse{}
	_, err := zk.request(opSetData, &setDataRequest{path, data, version}, res, nil)
	return &res.Stat, err
}

// 获取子节点
func (zk *ZK) children(path string) ([]string, error) {
	res := &getChildren2Response{}
	_, err := zk.request(opGetChildren2, &getChildren2Request{Path: path, Watch: false}, res, nil)
	return res.Children, err
}

// 新建一个节点
func (zk *ZK) create(path string, data []byte, acl []ACL, flags int32) (string, error) {
	res := &createResponse{}
	_, err := zk.request(opCreate, &createRequest{path, data, acl, flags}, res, nil)
	return res.Path, err
}

// 删除一个节点
func (zk *ZK) delete(path string, version int32) error {
	_, err := zk.request(opDelete, &deleteRequest{path, version}, &deleteResponse{}, nil)
	return err
}
