package main

import (
	"encoding/binary"
	"log"
	"net"
	"time"
)

const (
	RecvTimeout    = 1          // 接收消息超时，单位：秒
	SessionTimeout = 40000      // 客户端会话超时，单位：秒
	BufferSize     = 512 * 1024 // 512KB

	opCreate      = 1
	opDelete      = 2
	opExists      = 3
	opGetData     = 4
	opSetData     = 5
	opGetChildren = 8
	opPing        = 11
	opClose       = -11
)

var (
	// 测试用的IP
	TESTIP = []string{"192.168.56.101:2181"}

	// 第一次请求时，客户端发送16位空字节数组到服务器端
	emptyPassword = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

	// 全局ACL
	WorldACL = []ACL{{0x1f, "world", "anyone"}}
)

type ACL struct {
	Perms  int32
	Scheme string
	Id     string
}

type ZkCli struct {
	protocolversion int32
	sessiontimeout  int32
	sessionid       int64
	password        []byte
	conn            *net.TCPConn
}

func New() *ZkCli {
	zkCli := ZkCli{
		protocolversion: 0,
		sessiontimeout:  0,
		sessionid:       0,
		password:        emptyPassword,
		conn:            nil,
	}
	return &zkCli
}

func (zkCli *ZkCli) connect(serverAddr string) error {
	// TCP拔号
	tcpAddr, err := net.ResolveTCPAddr("tcp4", serverAddr)
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		log.Printf("[func connect] %v\n", err)
		return err
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
		log.Printf("[func connect] %v\n", err)
		return err
	}
	zkCli.conn.SetReadDeadline(time.Now().Add(RecvTimeout * time.Second))
	_, err = zkCli.conn.Read(buf)
	zkCli.conn.SetReadDeadline(time.Time{})
	if err != nil {
		log.Printf("[func connect] %v\n", err)
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
		log.Printf("[func close] %v\n", err)
		return err
	}
	err = zkCli.conn.Close()
	if err != nil {
		log.Printf("[func close] %v\n", err)
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
		log.Printf("[func ping] %v\n", err)
		return err
	}
	zkCli.conn.SetReadDeadline(time.Now().Add(RecvTimeout * time.Second))
	_, err = zkCli.conn.Read(buf)
	zkCli.conn.SetReadDeadline(time.Time{})
	if err != nil {
		log.Printf("[func ping] %v\n", err)
		return err
	}
	res := &pingResponse{}
	decodePingResponse(buf[4:], res)
	return nil
}

func (zkCli *ZkCli) exists(path string) (bool, error) {
	buf := make([]byte, 256)
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
		log.Printf("[func exist] %v\n", err)
		return false, err
	}
	zkCli.conn.SetReadDeadline(time.Now().Add(RecvTimeout * time.Second))
	_, err = zkCli.conn.Read(buf)
	zkCli.conn.SetReadDeadline(time.Time{})
	if err != nil {
		log.Printf("[func exist] %v\n", err)
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
	buf := make([]byte, 256)
	n := encodeGetRequest(buf, &getRequest{
		xid:    0,
		opcode: opGetData,
		path:   path,
		watch:  false,
	})
	zkCli.conn.SetWriteDeadline(time.Now().Add(RecvTimeout * time.Second))
	_, err := zkCli.conn.Write(buf[:n])
	zkCli.conn.SetWriteDeadline(time.Time{})
	if err != nil {
		log.Printf("[func get] %v\n", err)
		return nil, err
	}
	zkCli.conn.SetReadDeadline(time.Now().Add(RecvTimeout * time.Second))
	_, err = zkCli.conn.Read(buf)
	zkCli.conn.SetReadDeadline(time.Time{})
	if err != nil {
		log.Printf("[func get] %v\n", err)
		return nil, err
	}
	res := &getResponse{}
	decodeGetResponse(buf[4:], res)
	return res.data, nil
}

func (zkCli *ZkCli) children(path string) ([]string, error) {
	buf := make([]byte, 256)
	n := encodeChildrenRequest(buf, &childrenRequest{
		xid:    0,
		opcode: opGetChildren,
		path:   path,
		watch:  false,
	})
	zkCli.conn.SetWriteDeadline(time.Now().Add(RecvTimeout * time.Second))
	_, err := zkCli.conn.Write(buf[:n])
	zkCli.conn.SetWriteDeadline(time.Time{})
	if err != nil {
		log.Printf("[func children] %v\n", err)
		return nil, err
	}
	zkCli.conn.SetReadDeadline(time.Now().Add(RecvTimeout * time.Second))
	_, err = zkCli.conn.Read(buf)
	zkCli.conn.SetReadDeadline(time.Time{})
	if err != nil {
		log.Printf("[func children] %v\n", err)
		return nil, err
	}
	res := &childrenResponse{}
	decodeChildrenResponse(buf[4:], res)
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
		log.Printf("[func create] %v\n", err)
		return err
	}
	zkCli.conn.SetReadDeadline(time.Now().Add(RecvTimeout * time.Second))
	_, err = zkCli.conn.Read(buf)
	zkCli.conn.SetReadDeadline(time.Time{})
	if err != nil {
		log.Printf("[func create] %v\n", err)
		return err
	}
	res := &createResponse{}
	decodeCreateResponse(buf[4:], res)
	log.Println(res)
	return nil
}

func (zkCli *ZkCli) set(path string, data []byte) error {
	buf := make([]byte, BufferSize)
	n := encodeSetRequest(buf, &setRequest{
		xid:     0,
		opcode:  opSetData,
		path:    path,
		data:    data,
		version: -1, // 默认
	})
	zkCli.conn.SetWriteDeadline(time.Now().Add(RecvTimeout * time.Second))
	_, err := zkCli.conn.Write(buf[:n])
	zkCli.conn.SetWriteDeadline(time.Time{})
	if err != nil {
		log.Printf("[func set] %v\n", err)
		return err
	}
	zkCli.conn.SetReadDeadline(time.Now().Add(RecvTimeout * time.Second))
	_, err = zkCli.conn.Read(buf)
	zkCli.conn.SetReadDeadline(time.Time{})
	if err != nil {
		log.Printf("[func set] %v\n", err)
		return err
	}
	res := &setResponse{}
	decodeSetResponse(buf[4:], res)
	log.Println(res)
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
		log.Printf("[func delete] %v\n", err)
		return err
	}
	zkCli.conn.SetReadDeadline(time.Now().Add(RecvTimeout * time.Second))
	_, err = zkCli.conn.Read(buf)
	zkCli.conn.SetReadDeadline(time.Time{})
	if err != nil {
		log.Printf("[func delete] %v\n", err)
		return err
	}
	res := &deleteResponse{}
	decodeDeleteResponse(buf[4:], res)
	log.Println(res)
	return nil
}

// 测试
func (zkCli *ZkCli) Test(server string) {
	err := zkCli.connect(server)

	err = zkCli.ping()

	flag, err := zkCli.exists("/ymb")
	log.Println(flag, err)

	err = zkCli.close()

	err = zkCli.connect(server)
	data, err := zkCli.get("/ymb")
	log.Println(string(data), err)

	children, err := zkCli.children("/")
	log.Println(children, err)

	err = zkCli.create("/test2", []byte("goodgood"))
	log.Println(err)

	err = zkCli.set("/test", []byte("nonono"))
	log.Println(err)

	err = zkCli.delete("/test/okok")
	err = zkCli.delete("/test")
	log.Println(err)
}

func main() {
	zkCli := New()
	zkCli.Test(TESTIP[0])
}

func handleErr(err error) {
	if err != nil {
		log.Printf("%v\n", err)
		panic(err)
	}
}

/*----------------------------------------------
 *
 * 以下是各个协议的结构以及它们的编码和解码
 *
 ----------------------------------------------*/

// Header
type requestHeader struct {
	xid    int32 // 请求编号
	opcode int32
}

func encodeRequestHeader(buf []byte, req *requestHeader) int32 {
	copy(buf[0:], Int32ToBytes(req.xid))
	copy(buf[4:], Int32ToBytes(req.opcode))
	return 8
}

type responseHeader struct {
	xid     int32 // 响应编号
	zxid    int64
	errcode int32
}

func decodeResponseHeader(buf []byte, res *responseHeader) {
	res.xid = BytesToInt32(buf[0:])
	res.zxid = BytesToInt64(buf[4:])
	res.errcode = BytesToInt32(buf[12:])
}

// Connect
type connectRequest struct {
	protocolversion int32
	lastzxidseen    int64
	timeout         int32
	sessionid       int64
	password        []byte
}

func encodeConnectRequest(buf []byte, req *connectRequest) int32 {
	copy(buf[0:], Int32ToBytes(44))
	copy(buf[4:], Int32ToBytes(req.protocolversion))
	copy(buf[8:], Int64ToBytes(req.lastzxidseen))
	copy(buf[16:], Int32ToBytes(req.timeout))
	copy(buf[20:], Int64ToBytes(req.sessionid))
	copy(buf[28:], Int32ToBytes(16))
	copy(buf[32:], req.password)
	return 48
}

type connectResponse struct {
	protocolversion int32
	timeout         int32
	sessionid       int64
	password        []byte
}

func decodeConnectResponse(buf []byte, res *connectResponse) {
	res.protocolversion = BytesToInt32(buf[0:])
	res.timeout = BytesToInt32(buf[4:])
	res.sessionid = BytesToInt64(buf[8:])
	res.password = buf[16+4 : 16+4+16] // 对于slice，要先读一个四字节的长度，所以+4
}

// Close
type closeRequest struct {
	xid    int32
	opcode int32
}

func encodeCloseRequest(buf []byte, req *closeRequest) int32 {
	copy(buf[0:], Int32ToBytes(8))
	copy(buf[4:], Int32ToBytes(req.xid))
	return 8
}

type closeResponse struct {
	xid     int32
	zxid    int64
	errcode int32
}

func decodeCloseResponse(buf []byte, res *closeResponse) {
	res.xid = BytesToInt32(buf[0:])
	res.zxid = BytesToInt64(buf[4:])
	res.errcode = BytesToInt32(buf[12:])
}

// Ping
type pingRequest struct {
	xid    int32
	opcode int32
}

func encodePingRequest(buf []byte, req *pingRequest) int32 {
	copy(buf[0:], Int32ToBytes(8))
	copy(buf[4:], Int32ToBytes(req.xid))
	copy(buf[8:], Int32ToBytes(req.opcode))
	return 12
}

type pingResponse struct {
	xid     int32
	zxid    int64
	errcode int32
}

func decodePingResponse(buf []byte, res *pingResponse) {
	res.xid = BytesToInt32(buf[0:])
	res.zxid = BytesToInt64(buf[4:])
	res.errcode = BytesToInt32(buf[12:])
}

// Exist
type existRequest struct {
	xid    int32
	opcode int32
	path   string
	watch  bool
}

func encodeExistRequest(buf []byte, req *existRequest) int32 {
	length := int32(len(req.path))
	copy(buf[0:], Int32ToBytes(13+length))
	copy(buf[4:], Int32ToBytes(req.xid))
	copy(buf[8:], Int32ToBytes(req.opcode))
	copy(buf[12:], Int32ToBytes(length))
	copy(buf[16:], []byte(req.path))
	copy(buf[16+length:], BoolToBytes(req.watch))
	return 17 + length
}

type existResponse struct {
	xid     int32
	zxid    int64
	errcode int32
	// 这里还有节点的状态，但目前不关心
}

func decodeExistResponse(buf []byte, res *existResponse) {
	res.xid = BytesToInt32(buf[0:])
	res.zxid = BytesToInt64(buf[4:])
	res.errcode = BytesToInt32(buf[12:])
}

// Get
type getRequest struct {
	xid    int32
	opcode int32
	path   string
	watch  bool
}

func encodeGetRequest(buf []byte, req *getRequest) int32 {
	length := int32(len(req.path))
	copy(buf[0:], Int32ToBytes(13+length))
	copy(buf[4:], Int32ToBytes(req.xid))
	copy(buf[8:], Int32ToBytes(req.opcode))
	copy(buf[12:], Int32ToBytes(length))
	copy(buf[16:], []byte(req.path))
	copy(buf[16+length:], BoolToBytes(req.watch))
	return 17 + length
}

type getResponse struct {
	xid     int32
	zxid    int64
	errcode int32
	data    []byte
	// 这里还有节点的状态，但目前不关心
}

func decodeGetResponse(buf []byte, res *getResponse) {
	res.xid = BytesToInt32(buf[0:])
	res.zxid = BytesToInt64(buf[4:])
	res.errcode = BytesToInt32(buf[12:])
	length := BytesToInt32(buf[16:])
	res.data = buf[20 : 20+length]
}

// Children
type childrenRequest struct {
	xid    int32
	opcode int32
	path   string
	watch  bool
}

func encodeChildrenRequest(buf []byte, req *childrenRequest) int32 {
	length := int32(len(req.path))
	copy(buf[0:], Int32ToBytes(13+length))
	copy(buf[4:], Int32ToBytes(req.xid))
	copy(buf[8:], Int32ToBytes(req.opcode))
	copy(buf[12:], Int32ToBytes(length))
	copy(buf[16:], []byte(req.path))
	copy(buf[16+length:], BoolToBytes(req.watch))
	return 17 + length
}

type childrenResponse struct {
	xid      int32
	zxid     int64
	errcode  int32
	children []string
}

func decodeChildrenResponse(buf []byte, res *childrenResponse) {
	res.xid = BytesToInt32(buf[0:])
	res.zxid = BytesToInt64(buf[4:])
	res.errcode = BytesToInt32(buf[12:])
	children_cnt := BytesToInt32(buf[16:])
	var n int32 = 20
	for ; children_cnt > 0; children_cnt-- {
		child_len := BytesToInt32(buf[n:])
		n += 4
		res.children = append(res.children, string(buf[n:n+child_len]))
		n += child_len
	}
}

// Create
type createRequest struct {
	xid    int32
	opcode int32
	path   string
	data   []byte
	acl    []ACL
	flags  int32
}

func encodeCreateRequest(buf []byte, req *createRequest) int32 {
	path_len := len(req.path)
	data_len := len(req.data)
	acl_cnt := len(req.acl)
	n := 4
	copy(buf[n:], Int32ToBytes(req.xid))
	n += 4
	copy(buf[n:], Int32ToBytes(req.opcode))
	n += 4
	copy(buf[n:], Int32ToBytes(int32(path_len)))
	n += 4
	copy(buf[n:], []byte(req.path))
	n += path_len
	copy(buf[n:], Int32ToBytes(int32(data_len)))
	n += 4
	copy(buf[n:], req.data)
	n += data_len
	copy(buf[n:], Int32ToBytes(int32(acl_cnt)))
	n += 4
	for i := 0; i < acl_cnt; i++ {
		copy(buf[n:], Int32ToBytes(req.acl[i].Perms))
		n += 4
		acl_scheme_len := len(req.acl[i].Scheme)
		copy(buf[n:], Int32ToBytes(int32(acl_scheme_len)))
		n += 4
		copy(buf[n:], []byte(req.acl[i].Scheme))
		n += acl_scheme_len
		acl_id_len := len(req.acl[i].Id)
		copy(buf[n:], Int32ToBytes(int32(acl_id_len)))
		n += 4
		copy(buf[n:], []byte(req.acl[i].Id))
		n += acl_id_len
	}
	copy(buf[n:], Int32ToBytes(req.flags))
	copy(buf[:4], Int32ToBytes(int32(n)))
	return int32(n + 4)
}

type createResponse struct {
	xid     int32
	zxid    int64
	errcode int32
	// 这里还有节点的路径，但目前不关心
}

func decodeCreateResponse(buf []byte, res *createResponse) {
	res.xid = BytesToInt32(buf[0:])
	res.zxid = BytesToInt64(buf[4:])
	res.errcode = BytesToInt32(buf[12:])
}

// Set

type setRequest struct {
	xid     int32
	opcode  int32
	path    string
	data    []byte
	version int32
}

func encodeSetRequest(buf []byte, req *setRequest) int32 {
	path_len := len(req.path)
	data_len := len(req.data)
	n := 4
	copy(buf[n:], Int32ToBytes(req.xid))
	n += 4
	copy(buf[n:], Int32ToBytes(req.opcode))
	n += 4
	copy(buf[n:], Int32ToBytes(int32(path_len)))
	n += 4
	copy(buf[n:], []byte(req.path))
	n += path_len
	copy(buf[n:], Int32ToBytes(int32(data_len)))
	n += 4
	copy(buf[n:], req.data)
	n += data_len
	copy(buf[n:], Int32ToBytes(req.version))
	copy(buf[:4], Int32ToBytes(int32(n)))
	return int32(n + 4)
}

type setResponse struct {
	xid     int32
	zxid    int64
	errcode int32
	// 这里还有节点的状态，但目前不关心
}

func decodeSetResponse(buf []byte, res *setResponse) {
	res.xid = BytesToInt32(buf[0:])
	res.zxid = BytesToInt64(buf[4:])
	res.errcode = BytesToInt32(buf[12:])
}

// Delete

type deleteRequest struct {
	xid     int32
	opcode  int32
	path    string
	version int32
}

func encodeDeleteRequest(buf []byte, req *deleteRequest) int32 {
	path_len := len(req.path)
	n := 4
	copy(buf[n:], Int32ToBytes(req.xid))
	n += 4
	copy(buf[n:], Int32ToBytes(req.opcode))
	n += 4
	copy(buf[n:], Int32ToBytes(int32(path_len)))
	n += 4
	copy(buf[n:], []byte(req.path))
	n += path_len
	copy(buf[n:], Int32ToBytes(req.version))
	copy(buf[:4], Int32ToBytes(int32(n)))
	return int32(n + 4)
}

type deleteResponse struct {
	xid     int32
	zxid    int64
	errcode int32
}

func decodeDeleteResponse(buf []byte, res *deleteResponse) {
	res.xid = BytesToInt32(buf[0:])
	res.zxid = BytesToInt64(buf[4:])
	res.errcode = BytesToInt32(buf[12:])
}

/*----------------------------------------------
 *
 * 以下是基本类型的编码和解码
 *
 ----------------------------------------------*/

func Int32ToBytes(i int32) []byte {
	var buf = make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(i))
	return buf
}
func Int64ToBytes(i int64) []byte {
	var buf = make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(i))
	return buf
}
func BoolToBytes(b bool) []byte {
	var buf = make([]byte, 1)
	if b {
		buf[0] = 1
	} else {
		buf[0] = 0
	}
	return buf
}

func BytesToInt32(buf []byte) int32 {
	return int32(binary.BigEndian.Uint32(buf))
}
func BytesToInt64(buf []byte) int64 {
	return int64(binary.BigEndian.Uint64(buf))
}
func BytesToBool(buf []byte) bool {
	if buf[0] == 1 {
		return true
	}
	return false
}
