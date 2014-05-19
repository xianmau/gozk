package main

import (
	"fmt"
	"time"
	"zk"
	//"errors"
	//"encoding/binary"
	//"net"
	//"reflect"
	//"runtime"
)

func main() {
	fmt.Println("Zookeeper Client.")
	fmt.Println("-----------------")

	zk, err := zk.Connect([]string{"127.0.0.1"}, time.Second)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", zk)

	node := zk.Get("/")
	fmt.Printf("%+v\n", node)
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
