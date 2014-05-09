package main

import (
	"fmt"
	"time"
	//"zk"
	"net"
)

func main() {
	fmt.Println("Zookeeper Client.")
	//conn, _, err := zk.Connect([]string{"172.19.32.16", "172.19.32.153"}, time.Second)
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println(conn)

	zkConn, err := net.DialTimeout("tcp", "172.19.32.16:2181", time.Second)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", zkConn)
}

const (
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
	// 标准库里的TCP连接结构体
	conn net.Conn
	// 密码
	password []byte
}

func Connect(servers []string, connectTimeout time.Duration) (ZK, error) {
	zk := ZK{
		servers:        servers,
		connectTimeout: connectTimeout,
		sessionId:      0,
		conn:           nil,
		password:       emptyPassword,
	}
	conn, err := net.DialTimeout("tcp", "172.19.32.16:2181", connectTimeout)
	if err == nil {
		zk.conn = conn
	}
	return zk, err
}
func (zk *ZK) Get(path string) []byte {
	ret := []byte{"test"}
}
func (zk *ZK) Create(path string, data []byte) {

}
