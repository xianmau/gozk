package zk

import (
	"fmt"
	"strings"
	"time"
)

// API:连接
func Connect(servers []string, recvTimeout time.Duration) *ZK {
	// 处理没有带端口号的地址
	for index, serverip := range servers {
		if !strings.ContainsRune(serverip, rune(':')) {
			servers[index] = fmt.Sprintf("%s:%d", serverip, defaultPort)
		}
	}
	// 初始化一个实例
	// 然后客户端就一直用这个实例来对服务器进行访问
	// 也就是相当于客户端的全局变量
	zk := ZK{
		servers:           servers,                                  // 服务器地址集合
		serversIndex:      0,                                        // 连接到的服务器下标
		conn:              nil,                                      //
		connectTimeout:    1 * time.Second,                          // 连接超时为1秒
		sessionId:         0,                                        // 会话Id，第一次连接会重服务端获取
		sessionTimeout:    86400,                                    // 会话超时为一天，呵呵
		password:          emptyPassword,                            // 密码
		state:             StateDisconnected,                        // 连接状态
		heartbeatInterval: time.Duration((int64(recvTimeout) >> 1)), // 心跳周期，为接收超时的一半
		recvTimeout:       recvTimeout,                              // 接收超时
		shouldQuit:        make(chan bool),                          //
		sendChan:          make(chan *request, sendChanSize),        // 消息队列，队列里的每个消息为一个请求
		requests:          make(map[int32]*request),                 // 请求映射
	}
	// 开个协程来连接
	go func() {
		zk.connect(servers, recvTimeout)
	}()
	return &zk
}

func (zk *ZK) Close() {
	zk.close()
}

// 测试节点是否存在
func (zk *ZK) Exists(path string) (bool, error) {
	flag, _, err := zk.exists(path)
	return flag, err
}

// 获取节点数据
func (zk *ZK) Get(path string) (string, error) {
	data, _, err := zk.get(path)
	return string(data), err
}

// 设置节点数据
func (zk *ZK) Set(path string, data string) error {
	_, err := zk.set(path, []byte(data), -1)
	return err
}

// 获取子节点列表
func (zk *ZK) Children(path string) ([]string, error) {
	return zk.children(path)
}

// 新建
func (zk *ZK) Create(path string, data string, acl []ACL, flags int32) error {
	_, err := zk.create(path, []byte(data), acl, flags)
	return err
}

// 删除
func (zk *ZK) Delete(path string) error {
	return zk.delete(path, -1)
}

// 递归删除
func (zk *ZK) DeleteRecur(path string) error {
	if flag, err := zk.Exists(path); err == nil && !flag {
		return err
	}
	children, err := zk.Children(path)
	if err != nil {
		return err
	}
	for _, znode := range children {
		sub_znode := path + "/" + znode
		zk.DeleteRecur(sub_znode)
	}
	return zk.Delete(path)
}
