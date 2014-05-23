package xxx

/*
TODO:
* make sure a ping response comes back in a reasonable time

Possible watcher events:
* Event{Type: EventNotWatching, State: StateDisconnected, Path: path, Err: err}
*/

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	bufferSize      = 1536 * 1024
	eventChanSize   = 6
	sendChanSize    = 16
	protectedPrefix = "_c_"
)

type watchType int

const (
	watchTypeData  = iota
	watchTypeExist = iota
	watchTypeChild = iota
)

type watchPathType struct {
	path  string
	wType watchType
}

type Dialer func(network, address string, timeout time.Duration) (net.Conn, error)

type Conn struct {
	lastZxid  int64
	sessionID int64
	state     State // must be 32-bit aligned
	xid       int32
	timeout   int32 // session timeout in seconds
	passwd    []byte

	dialer         Dialer
	servers        []string
	serverIndex    int
	conn           net.Conn
	eventChan      chan Event
	shouldQuit     chan bool
	pingInterval   time.Duration
	recvTimeout    time.Duration
	connectTimeout time.Duration

	sendChan     chan *request
	requests     map[int32]*request // Xid -> pending request
	requestsLock sync.Mutex
	watchers     map[watchPathType][]chan Event
	watchersLock sync.Mutex

	// Debug (used by unit tests)
	reconnectDelay time.Duration
}

type request struct {
	xid        int32
	opcode     int32
	pkt        interface{}
	recvStruct interface{}
	recvChan   chan response

	// Because sending and receiving happen in separate go routines, there's
	// a possible race condition when creating watches from outside the read
	// loop. We must ensure that a watcher gets added to the list synchronously
	// with the response from the server on any request that creates a watch.
	// In order to not hard code the watch logic for each opcode in the recv
	// loop the caller can use recvFunc to insert some synchronously code
	// after a response.
	recvFunc func(*request, *responseHeader, error)
}

type response struct {
	zxid int64
	err  error
}

type Event struct {
	Type  EventType
	State State
	Path  string // For non-session events, the path of the watched node.
	Err   error
}

// 连接
func Connect(servers []string, recvTimeout time.Duration) (*Conn, <-chan Event, error) {
	return ConnectWithDialer(servers, recvTimeout, nil)
}

func ConnectWithDialer(servers []string, recvTimeout time.Duration, dialer Dialer) (*Conn, <-chan Event, error) {
	conn := Conn{
		dialer:         dialer,
		servers:        servers,
		serverIndex:    0, // 标记当前连接的是哪个节点
		conn:           nil,
		state:          StateDisconnected,
		eventChan:      ec,
		shouldQuit:     make(chan bool),
		recvTimeout:    recvTimeout,
		pingInterval:   time.Duration((int64(recvTimeout) / 2)), // 心跳时间为超时的一半
		connectTimeout: 1 * time.Second,
		sendChan:       make(chan *request, sendChanSize),
		requests:       make(map[int32]*request),
		watchers:       make(map[watchPathType][]chan Event),
		passwd:         emptyPassword,
		timeout:        30000,

		// Debug
		reconnectDelay: 0,
	}
	go func() { // 开一个协程处理
		conn.loop() // loop的意思可能是不断地尝试那些节点，因为有可能某些节点挂了
		conn.flushRequests(ErrClosing)
		conn.invalidateWatches(ErrClosing)
	}()
	return &conn, ec, nil
}

func (c *Conn) connect() {
	c.serverIndex = (c.serverIndex + 1) % len(c.servers)
	startIndex := c.serverIndex
	c.setState(StateConnecting) // 设置状态为：正在连接
	for {
		zkConn, err := c.dialer("tcp", c.servers[c.serverIndex], c.connectTimeout) // 进行TCP拔号
		if err == nil {
			c.conn = zkConn
			c.setState(StateConnected) // 设置状态为：已连接
			return
		}

		log.Printf("Failed to connect to %s: %+v", c.servers[c.serverIndex], err) // 写一条日志没啥

		// 如果失败进行尝试下一个节点
		c.serverIndex = (c.serverIndex + 1) % len(c.servers)
		if c.serverIndex == startIndex { // 这里表明已经尝试了一圈节点都没成功，于是，先休眠一下
			time.Sleep(time.Second)
		}
	}
}

func (c *Conn) authenticate() error {
	buf := make([]byte, 256) // 这是一个缓冲区

	// connect request

	// 对连接请求编码成字节数组
	n, err := encodePacket(buf[4:], &connectRequest{
		ProtocolVersion: protocolVersion,
		LastZxidSeen:    c.lastZxid,
		TimeOut:         c.timeout,
		SessionID:       c.sessionID,
		Passwd:          c.passwd,
	})
	if err != nil {
		return err
	}

	binary.BigEndian.PutUint32(buf[:4], uint32(n)) // 在字节数组前4个字节标明了请求的长度

	_, err = c.conn.Write(buf[:n+4]) // conn.Write其实就是向服务器发数据
	if err != nil {
		return err
	}

	// 设置监视，其实先不考虑监视了
	//c.sendSetWatches()

	// connect response

	// package length
	_, err = io.ReadFull(c.conn, buf[:4]) // 读取包的大小
	if err != nil {
		return err
	}

	blen := int(binary.BigEndian.Uint32(buf[:4])) // 转成数字，然后判断缓冲区够不够大
	if cap(buf) < blen {
		buf = make([]byte, blen)
	}

	_, err = io.ReadFull(c.conn, buf[:blen]) // 读取数据
	if err != nil {
		return err
	}

	// 把读到的数据解码成结构
	r := connectResponse{}
	_, err = decodePacket(buf[:blen], &r)
	if err != nil {
		return err
	}
	if r.SessionID == 0 {
		c.sessionID = 0
		c.passwd = emptyPassword
		c.lastZxid = 0
		c.setState(StateExpired) // 设置状态为：超时
		return ErrSessionExpired // 返回超时错误
	}

	if c.sessionID != r.SessionID {
		//客户端和服务器在交互时，有几个特殊的xid，分别的表示含义如下：
		//Request: xid: -8表示reconnect时重新设置watches, -2表示ping 包，-4表示auth
		//Response: xid -2 表示ping的响应，-4 表示auth的响应，-1表示本次响应时有watch事件发生
		atomic.StoreInt32(&c.xid, 0) // 其实就是互斥写操作
	}
	c.timeout = r.TimeOut
	c.sessionID = r.SessionID
	c.passwd = r.Passwd
	c.setState(StateHasSession) // 设置状态：已有连接

	return nil
}

func (c *Conn) loop() {
	for {
		c.connect()             // 连接
		err := c.authenticate() // 进行认证
		switch {
		// 会话超时，先不理了
		//case err == ErrSessionExpired:
		//	c.invalidateWatches(err)
		case err != nil && c.conn != nil: // 如果出错了，关闭TCP连接
			c.conn.Close()
		case err == nil:
			// 认证成功的话，通知seenloop停止
			closeChan := make(chan bool) // channel to tell send loop stop

			var wg sync.WaitGroup // 用于同步线程

			wg.Add(1) // 先做完这个
			go func() {
				c.sendLoop(c.conn, closeChan) // 包括发心跳和其它请求到服务器
				c.conn.Close()                // causes recv loop to EOF/exit
				wg.Done()
			}()

			wg.Add(1) // 再做这个
			go func() {
				err = c.recvLoop(c.conn) // 应该就是读取数据
				if err == nil {
					panic("zk: recvLoop should never return nil error")
				}
				close(closeChan) // tell send loop to exit
				wg.Done()
			}()

			wg.Wait()
		}

		// 这是的意思应该是上面结束了，然后就断开连接，准备等下重新连吧
		c.setState(StateDisconnected) // 设置状态为：未连接

		// Yeesh
		if err != io.EOF && err != ErrSessionExpired && !strings.Contains(err.Error(), "use of closed network connection") {
			log.Println(err)
		}

		select {
		case <-c.shouldQuit:
			c.flushRequests(ErrClosing)
			return
		default:
		}

		if err != ErrSessionExpired {
			err = ErrConnectionClosed
		}
		c.flushRequests(err)

		// 重连延迟
		if c.reconnectDelay > 0 {
			select {
			case <-c.shouldQuit:
				return
			case <-time.After(c.reconnectDelay): // 阻塞一段时间
			}
		}
	}
}

// Send error to all pending requests and clear request map
func (c *Conn) flushRequests(err error) {
	c.requestsLock.Lock()
	for _, req := range c.requests {
		req.recvChan <- response{-1, err}
	}
	c.requests = make(map[int32]*request)
	c.requestsLock.Unlock()
}

func (c *Conn) sendLoop(conn net.Conn, closeChan <-chan bool) error {
	pingTicker := time.NewTicker(c.pingInterval) // 设置心跳间隔，一般是接收超时的一半时间
	defer pingTicker.Stop()                      // 停止心跳

	buf := make([]byte, bufferSize)
	for {
		select {
		case req := <-c.sendChan: // sendChan是请求队列，这里获取一个请求然后执行
			header := &requestHeader{req.xid, req.opcode} // 请求包头
			n, err := encodePacket(buf[4:], header)       // 对请求包头进行编码
			if err != nil {
				req.recvChan <- response{-1, err} // 如果呢编码失败，那么request结构体中的recvChan就加入一个错误的响应，然后跳过这次循环
				continue
			}

			n2, err := encodePacket(buf[4+n:], req.pkt) // 对请求包编码成字节数组
			if err != nil {
				req.recvChan <- response{-1, err} // 如果呢编码失败，那么request结构体中的recvChan就加入一个错误的响应，然后跳过这次循环
				continue
			}

			n += n2 // 请求包头和请求包的总长度

			binary.BigEndian.PutUint32(buf[:4], uint32(n))

			c.requestsLock.Lock()
			// 感觉有select就是为了把线程阻塞的样子
			select {
			case <-closeChan:
				req.recvChan <- response{-1, ErrConnectionClosed}
				c.requestsLock.Unlock()
				return ErrConnectionClosed
			default:
			}
			c.requests[req.xid] = req
			c.requestsLock.Unlock()

			conn.SetWriteDeadline(time.Now().Add(c.recvTimeout)) // 设置写超时
			_, err = conn.Write(buf[:n+4])                       // 相当于发送TCP消息
			conn.SetWriteDeadline(time.Time{})
			if err != nil {
				req.recvChan <- response{-1, err}
				conn.Close()
				return err
			}
		case <-pingTicker.C:
			n, err := encodePacket(buf[4:], &requestHeader{Xid: -2, Opcode: opPing})
			if err != nil {
				panic("zk: opPing should never fail to serialize")
			}

			binary.BigEndian.PutUint32(buf[:4], uint32(n))

			conn.SetWriteDeadline(time.Now().Add(c.recvTimeout))
			_, err = conn.Write(buf[:n+4])
			conn.SetWriteDeadline(time.Time{}) // 让它超时
			if err != nil {
				conn.Close()
				return err
			}
		case <-closeChan: // 取出来的如果是true，那么就结束喽
			return nil
		}
	}
}

func (c *Conn) recvLoop(conn net.Conn) error {
	buf := make([]byte, bufferSize)
	for {
		// package length
		conn.SetReadDeadline(time.Now().Add(c.recvTimeout)) // 设置读超时
		_, err := io.ReadFull(conn, buf[:4])                // 先读长度
		if err != nil {
			return err
		}

		blen := int(binary.BigEndian.Uint32(buf[:4]))
		if cap(buf) < blen {
			buf = make([]byte, blen)
		}

		_, err = io.ReadFull(conn, buf[:blen]) // 读数据
		conn.SetReadDeadline(time.Time{})      // 设置它超时
		if err != nil {
			return err
		}

		res := responseHeader{}
		_, err = decodePacket(buf[:16], &res) // 对前16个字节进行解码，其中Xid占4字节，Zxid占8字节，Err占4字节
		if err != nil {
			return err
		}

		// 接下来就根据这个Xid进行各种相应的操作了
		if res.Xid == -1 {
			res := &watcherEvent{}
			_, err := decodePacket(buf[16:16+blen], res)
			if err != nil {
				return err
			}
			ev := Event{
				Type:  res.Type,
				State: res.State,
				Path:  res.Path,
				Err:   nil,
			}
			select {
			case c.eventChan <- ev:
			default:
			}
			wTypes := make([]watchType, 0, 2)
			switch res.Type {
			case EventNodeCreated:
				wTypes = append(wTypes, watchTypeExist)
			case EventNodeDeleted, EventNodeDataChanged:
				wTypes = append(wTypes, watchTypeExist, watchTypeData, watchTypeChild)
			case EventNodeChildrenChanged:
				wTypes = append(wTypes, watchTypeChild)
			}
			c.watchersLock.Lock()
			for _, t := range wTypes {
				wpt := watchPathType{res.Path, t}
				if watchers := c.watchers[wpt]; watchers != nil && len(watchers) > 0 {
					for _, ch := range watchers {
						ch <- ev
						close(ch)
					}
					delete(c.watchers, wpt)
				}
			}
			c.watchersLock.Unlock()
		} else if res.Xid == -2 {
			// Ping response. Ignore.
		} else if res.Xid < 0 {
			log.Printf("Xid < 0 (%d) but not ping or watcher event", res.Xid)
		} else {
			if res.Zxid > 0 {
				c.lastZxid = res.Zxid
			}

			c.requestsLock.Lock()
			req, ok := c.requests[res.Xid]
			if ok {
				delete(c.requests, res.Xid) // 删除已经响应的请求
			}
			c.requestsLock.Unlock()

			if !ok {
				log.Printf("Response for unknown request with xid %d", res.Xid) // 表示得到的响应不知道是哪个请求的结果
			} else {
				if res.Err != 0 {
					err = res.Err.toError()
				} else {
					_, err = decodePacket(buf[16:16+blen], req.recvStruct) // 解码
				}
				if req.recvFunc != nil {
					req.recvFunc(req, &res, err) // 这个是回调函数，暂时不理它
				}
				req.recvChan <- response{res.Zxid, err} // 把接收的数据存起来
				if req.opcode == opClose {              // 如果收到的结果代号是退出的代号则退出
					return io.EOF
				}
			}
		}
	}
}

// 感觉这个就是来对应消息队列里的顺序
func (c *Conn) nextXid() int32 {
	return atomic.AddInt32(&c.xid, 1)
}

func (c *Conn) queueRequest(opcode int32, req interface{}, res interface{}, recvFunc func(*request, *responseHeader, error)) <-chan response {
	rq := &request{
		xid:        c.nextXid(),
		opcode:     opcode,
		pkt:        req,
		recvStruct: res,
		recvChan:   make(chan response, 1),
		recvFunc:   recvFunc,
	}
	c.sendChan <- rq
	return rq.recvChan
}

func (c *Conn) request(opcode int32, req interface{}, res interface{}, recvFunc func(*request, *responseHeader, error)) (int64, error) {
	r := <-c.queueRequest(opcode, req, res, recvFunc)
	return r.zxid, r.err
}

func (c *Conn) AddAuth(scheme string, auth []byte) error {
	_, err := c.request(opSetAuth, &setAuthRequest{Type: 0, Scheme: scheme, Auth: auth}, &setAuthResponse{}, nil)
	return err
}

func (c *Conn) Children(path string) ([]string, *Stat, error) {
	res := &getChildren2Response{}
	_, err := c.request(opGetChildren2, &getChildren2Request{Path: path, Watch: false}, res, nil)
	return res.Children, &res.Stat, err
}

func (c *Conn) Get(path string) ([]byte, *Stat, error) {
	res := &getDataResponse{}
	_, err := c.request(opGetData, &getDataRequest{Path: path, Watch: false}, res, nil)
	return res.Data, &res.Stat, err
}

func (c *Conn) Set(path string, data []byte, version int32) (*Stat, error) {
	res := &setDataResponse{}
	_, err := c.request(opSetData, &SetDataRequest{path, data, version}, res, nil)
	return &res.Stat, err
}

func (c *Conn) Create(path string, data []byte, flags int32, acl []ACL) (string, error) {
	res := &createResponse{}
	_, err := c.request(opCreate, &CreateRequest{path, data, acl, flags}, res, nil)
	return res.Path, err
}

func (c *Conn) Delete(path string, version int32) error {
	_, err := c.request(opDelete, &DeleteRequest{path, version}, &deleteResponse{}, nil)
	return err
}

func (c *Conn) Exists(path string) (bool, *Stat, error) {
	res := &existsResponse{}
	_, err := c.request(opExists, &existsRequest{Path: path, Watch: false}, res, nil)
	exists := true
	if err == ErrNoNode {
		exists = false
		err = nil
	}
	return exists, &res.Stat, err
}

func (c *Conn) GetACL(path string) ([]ACL, *Stat, error) {
	res := &getAclResponse{}
	_, err := c.request(opGetAcl, &getAclRequest{Path: path}, res, nil)
	return res.Acl, &res.Stat, err
}

func (c *Conn) SetACL(path string, acl []ACL, version int32) (*Stat, error) {
	res := &setAclResponse{}
	_, err := c.request(opSetAcl, &setAclRequest{Path: path, Acl: acl, Version: version}, res, nil)
	return &res.Stat, err
}

func (c *Conn) Sync(path string) (string, error) {
	res := &syncResponse{}
	_, err := c.request(opSync, &syncRequest{Path: path}, res, nil)
	return res.Path, err
}

func (c *Conn) Close() {
	close(c.shouldQuit)

	select {
	case <-c.queueRequest(opClose, &closeRequest{}, &closeResponse{}, nil):
	case <-time.After(time.Second):
	}
}

func (c *Conn) State() State {
	return State(atomic.LoadInt32((*int32)(&c.state)))
}

func (c *Conn) setState(state State) {
	atomic.StoreInt32((*int32)(&c.state), int32(state))
	select {
	case c.eventChan <- Event{Type: EventSession, State: state}:
	default:
		// panic("zk: event channel full - it must be monitored and never allowed to be full")
	}
}
