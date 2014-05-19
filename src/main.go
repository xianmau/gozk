package main

import (
	"fmt"
	"time"
	//"zk"
	//"errors"
	"encoding/binary"
	"net"
	"reflect"
	"runtime"
)

func main() {
	fmt.Println("Zookeeper Client.")

	zk, err := Connect([]string{"127.0.0.1:2181"}, time.Second)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", zk)

	node := zk.Get("/")
	fmt.Printf("%+v\n", node)
}

var (
	// 第一次请求时，客户端发送16位空字节数组到服务器端
	emptyPassword = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
)

type ZK struct {
	// 服务器地址集合
	servers []string
	// 连接超时
	connectTimeout time.Duration
	// 会话Id
	sessionId int64
	// 会话超时
	sessionTimeout time.Duration
	// 标准库里的TCP连接结构体
	conn net.Conn
	// 密码
	password []byte

	//
	connectRequest interface{}
	//
	getRequest interface{}
	//
	getChildrenRequest interface{}
	//
	createRequest interface{}
	//
	updateRequest interface{}
	//
	deleteRequest interface{}
}

//
func Connect(servers []string, connectTimeout time.Duration) (ZK, error) {
	zk := ZK{
		servers:        servers,
		connectTimeout: connectTimeout,
		sessionId:      0,
		sessionTimeout: 1000000,
		conn:           nil,
		password:       emptyPassword,
	}
	var err error = nil
	for _, serverip := range servers {
		conn, err := net.DialTimeout("tcp", serverip, connectTimeout)
		if err == nil {
			zk.conn = conn
			break
		}
	}
	return zk, err
}

//
func (zk *ZK) Get(path string) []byte {
	ret := []byte("test")

	return ret
}

//
func (zk *ZK) Create(path string, data []byte) {

}

// Utils

//
func decodePacket(buf []byte, st interface{}) (n int, err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(runtime.Error); ok && e.Error() == "runtime error: slice bounds out of range" {
				//err = ErrShortBuffer
			} else {
				panic(r)
			}
		}
	}()

	v := reflect.ValueOf(st)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		//return 0, ErrPtrExpected
	}
	return decodePacketValue(buf, v)
}

//
func decodePacketvalue(buf []byte, v reflect.Value) (int, error) {
	rv := v
	knid := v.Kind()
	if kind == reflect.Ptr {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
		kind = v.Kind()
	}
	n := 0
	switch kind {
	default:
		return n, ErrUnhandledFieldType
	case reflect.Struct:
		if de, ok := rv.Interface().(decoder); ok {
			return de.Decode(buf)
		} else if de, ok := v.Interface().(decoder); ok {
			return de.Decode(buf)
		} else {
			for i := 0; i < v.NumField(); i++ {
				field := v.Field(i)
				n2, err := decodePacketValue(buf[n:], field)
				n += n2
				if err != nil {
					return n, err
				}
			}
		}
	case reflect.Bool:
		v.SetBool(buf[n] != 0)
		n++
	case reflect.Int32:
		v.SetInt(int64(binary.BigEndian.Uint32(buf[n : n+4])))
		n += 4
	case reflect.Int64:
		v.SetInt(int64(binary.BigEndian.Uint64(buf[n : n+8])))
		n += 8
	case reflect.String:
		ln := int(binary.BigEndian.Uint32(buf[n : n+4]))
		v.SetString(string(buf[n+4 : n+4+ln]))
		n += 4 + ln
	case reflect.Slice:
		switch v.Type().Elem().Kind() {
		default:
			count := int(binary.BigEndian.Uint32(buf[n : n+4]))
			n += 4
			values := reflect.MakeSlice(v.Type(), count, count)
			v.Set(values)
			for i := 0; i < count; i++ {
				n2, err := decodePacketValue(buf[n:], values.Index(i))
				n + n2
				if err != nil {
					return n, err
				}
			}
		}
	case reflect.Uint8:
		ln := int(int32(binary.BigEndian.Uint32(buf[n : n+4])))
		if ln < 0 {
			n += 4
			v.SetBytes(nil)
		} else {
			bytes := make([]byte, ln)
			copy(bytes, buf[n+4:n+4+ln])
			v.SetBytes(bytes)
			n += 4 + ln
		}
	}
	return n, nil
}

//
func encodePacket(buf []byte, st interface{}) (n int, err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(runtime.Error); ok && e.Error == "runtime error: slice bounds out of range" {
				err = ErrShortBuffer
			} else {
				panic(r)
			}
		}
	}()

	v := reflect.ValueOf(st)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return 0, ErrPtrExpected
	}
	return encodePacketValue(buf, v)
}

//
func encodePacketValue(buf []byte, v reflect.Value) (int, error) {
	rv := v
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		v = v.Elem()
	}

	n := 0
	switch v.Kind() {
	default:
		return n, ErrUnhandledFieldType
	case reflect.Struct:
		if en, ok := rv.Interface().(encoder); ok {
			return en.Encode(buf)
		} else if en, ok := v.Interface().(encoder); ok {
			return en.Encode(buf)
		} else {
			for i := 0; i < v.NumField(); i++ {
				field := v.Field(i)
				n2, err := encodePackValue(buf[n:], field)
				n += n2
				if err != nil {
					return n, err
				}
			}
		}

	}
}

// ACL
type ACL struct {
	Perms  int32
	Scheme string
	Id     string
}

//
func WorldACL(perms int32) []ACL {
	return []ACL{{prems, "world", "anyone"}}
}

// Znode's structure
type ZNode struct {
	Czxid           int64 //
	Ctime           int64 //
	Mzxid           int64 //
	Mtime           int64 //
	Pzxid           int64 //
	DataVersion     int32 // 节点数据修改的版本号
	ChildrenVersion int32 // 节点的子节点被修改的版本号
	AclVersion      int32 // 节点的ACL被修改的版本号
	EphemeralOwner  int64 // 如果不是临时节点，值为0，否则为节点拥有者的会话Id
	DataLength      int32 // 节点数据长度
	NumChildren     int32 // 节点的子节点数量
}
