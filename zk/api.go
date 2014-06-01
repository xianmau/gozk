package zk

import (
	"fmt"
	"strings"
	"time"
)

// API：连接
func Connect(servers []string, recvTimeout time.Duration) *ZK {
	for index, serverip := range servers {
		if !strings.ContainsRune(serverip, rune(':')) {
			servers[index] = fmt.Sprintf("%s:%d", serverip, defaultPort)
		}
	}
	zk := ZK{
		servers:           servers, // 服务器地址集合
		serversIndex:      0, // 连接到的服务器下标
		conn:              nil, //
		connectTimeout:    1 * time.Second, // 连接超时为1秒
		sessionId:         0, // 会话Id，第一次连接会重服务端获取
		sessionTimeout:    86400, // 会话超时为一天，呵呵
		password:          emptyPassword, // 密码
		state:             StateDisconnected, // 连接状态
		heartbeatInterval: time.Duration((int64(recvTimeout) >> 1)), // 心跳周期，为接收超时的一半
		recvTimeout:       recvTimeout, // 接收超时
		shouldQuit:        make(chan bool), //
		sendChan:          make(chan *request, sendChanSize), // 消息队列，队列里的每个消息为一个请求
		requests:          make(map[int32]*request), // 请求映射
		reconnectDelay:    5 * time.Second, // 延迟重连的时间
	}
	go func() {
		zk.connect(servers, recvTimeout)
	}()
	return &zk
}

// API：关闭连接
func (zk *ZK) Close() {
	zk.close()
}

// API：测试节点是否存在
func (zk *ZK) Exists(path string) (bool, error) {
	flag, _, err := zk.exists(path)
	return flag, err
}

// API：获取节点数据
func (zk *ZK) Get(path string) (string, error) {
	data, _, err := zk.get(path)
	return string(data), err
}

// API：设置节点数据
func (zk *ZK) Set(path string, data string) error {
	_, err := zk.set(path, []byte(data), -1)
	return err
}

// API：获取子节点列表
func (zk *ZK) Children(path string) ([]string, error) {
	return zk.children(path)
}

// API：新建节点
func (zk *ZK) Create(path string, data string, acl []ACL, flags int32) error {
	_, err := zk.create(path, []byte(data), acl, flags)
	return err
}

// API：删除节点
func (zk *ZK) Delete(path string) error {
	return zk.delete(path, -1)
}

// API：递归删除节点
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
