package main

import (
	"encoding/binary"
	"log"
	"net"
	"time"
)

const (
	BufferSize = 1024 * 1024 // 1M

	opCreate       = 1
	opDelete       = 2
	opExists       = 3
	opGetData      = 4
	opSetData      = 5
	opGetChildren  = 8
	opPing         = 11
	opGetChildren2 = 12
	opClose        = -11
)

var (
	// 测试用的IP
	TESTIP = []string{
		"192.168.56.101:2181",
	}

	// 第一次请求时，客户端发送16位空字节数组到服务器端
	emptyPassword = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
)

func main() {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", TESTIP[0])
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	handleErr(err)
	if err == nil { // 连接成功那么进行处理
		buf := make([]byte, BufferSize)
		EncodeConnectRequest(buf, &connectRequest{
			protocolversion: 0,
			lastzxidseen:    0,
			timeout:         3000,
			sessionid:       0,
			password:        emptyPassword,
		})
		log.Printf("%v\n", buf[:44])
		conn.SetWriteDeadline(time.Now().Add(time.Second))
		getlen, err := conn.Write(buf[:44])
		conn.SetWriteDeadline(time.Time{})

		EncodePingRequest(buf, &pingRequest{
			xid:    -2,
			opcode: opPing,
		})
		conn.SetWriteDeadline(time.Now().Add(time.Second))
		getlen, err = conn.Write(buf[:44])
		conn.SetWriteDeadline(time.Time{})

		conn.SetReadDeadline(time.Now().Add(time.Second))
		getlen, err = conn.Read(buf)
		conn.SetReadDeadline(time.Time{})

		conn.SetReadDeadline(time.Now().Add(time.Second))
		getlen, err = conn.Read(buf)
		conn.SetReadDeadline(time.Time{})

		log.Println(getlen, err, buf[:20])
	}
	log.Printf("%v", conn)
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
	xid    int32 // -2表示ping包，-4表示auth包，-8表示重连时重新设置watches
	opcode int32
}

func EncodeRequestHeader(buf []byte, req *requestHeader) int32 {
	copy(buf[0:], Int32ToBytes(req.xid))
	copy(buf[4:], Int32ToBytes(req.opcode))
	return 8
}

type responseHeader struct {
	xid     int32 // -2表示ping的响应，-4表示auth的响应，-1表示本次响应有watch事件发生
	zxid    int64
	errcode int32
}

func DecodeResponseHeader(buf []byte, res *responseHeader) {
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

func EncodeConnectRequest(buf []byte, req *connectRequest) int32 {
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

func DecodeConnectResponse(buf []byte, res *connectResponse) {
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

func EncodeCloseRequest(buf []byte, req *closeRequest) int32 {
	copy(buf[0:], Int32ToBytes(8))
	copy(buf[4:], Int32ToBytes(req.xid))
	return 8
}

type closeResponse struct{}

// Ping
type pingRequest struct {
	xid    int32
	opcode int32
}

func EncodePingRequest(buf []byte, req *pingRequest) int32 {
	copy(buf[0:], Int32ToBytes(8))
	copy(buf[4:], Int32ToBytes(req.xid))
	copy(buf[8:], Int32ToBytes(req.opcode))
	return 12
}

type pingResponse struct{}

// Exist
type existRequest struct {
	xid    int32
	opcode int32
	path   string
	watch  bool
}

func EncodeExistRequest(buf []byte, req *existRequest) int32 {
	l := len(req.path)
	copy(buf[0:], Int32ToBytes(8+))
	copy(buf[4:], Int32ToBytes(req.xid))
	copy(buf[8:], Int32ToBytes(req.opcode))
	copy(buf[12:], Int32ToBytes(l))
	copy(buf[16:], []byte(req.path))
	copy(buf[16+l:], BoolToBytes())
	return 12
}

// 基本类型编解码
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
func BytesToInt32(buf []byte) int32 {
	return int32(binary.BigEndian.Uint32(buf))
}
func BytesToInt64(buf []byte) int64 {
	return int64(binary.BigEndian.Uint64(buf))
}
func BoolToBytes(b bool) []byte{
	var buf = make([]byte,1)
	if b {
		buf[0] = 1
	}else{
		buf[0] = 0
	}
	return buf
}
