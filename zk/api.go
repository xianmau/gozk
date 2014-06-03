package zk

import (
	"fmt"
	"strings"
)

// API：新建实例
func New() *ZkCli {
	zkCli := ZkCli{
		protocolversion: 0,
		sessiontimeout:  0,
		sessionid:       0,
		password:        []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		conn:            nil,
		state:           stateDisconnect,
	}
	return &zkCli
}

// API：连接
func (zk *ZkCli) Connect(servers []string) error {
	for index, serverip := range servers {
		if !strings.ContainsRune(serverip, rune(':')) {
			servers[index] = fmt.Sprintf("%s:%d", serverip, DefaultPort)
		}
	}
	var err error = nil
	for index := range servers {
		err = zk.connect(servers[index])
		if err == nil {
			return nil
		}
	}
	logger.Println(err)
	return err
}

// API：Ping
func (zk *ZkCli) Ping() error {
	return zk.ping()
}

// API：关闭连接
func (zk *ZkCli) Close() error {
	return zk.close()
}

// API：测试节点是否存在
func (zk *ZkCli) Exists(path string) (bool, error) {
	return zk.exists(path)
}

// API：获取节点数据
func (zk *ZkCli) Get(path string) ([]byte, error) {
	return zk.get(path)
}

// API：获取子节点列表
func (zk *ZkCli) Children(path string) ([]string, error) {
	return zk.children(path)
}

// API：新建节点
func (zk *ZkCli) Create(path string, data []byte) error {
	return zk.create(path, data)
}

// API：设置节点数据
func (zk *ZkCli) Set(path string, data []byte) error {
	return zk.set(path, data)
}

// API：删除节点
func (zk *ZkCli) Delete(path string) error {
	return zk.delete(path)
}

// API：递归删除节点
func (zk *ZkCli) DeleteRecur(path string) error {
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
