package main

import (
	"fmt"
	"time"
	"zk"
	//"net"
)

var (
	TESTIP = []string{
		"172.19.32.39",
		//"192.168.56.101",
	}
)

func main() {
	fmt.Println("Zookeeper Client Start.")
	fmt.Println("-----------------------")

	z := zk.Connect(TESTIP, time.Second)

	time.Sleep(time.Second)
	fmt.Printf("## [zk] %+v\n", z)

	path, _ := z.Create("/test2", []byte("test data"), zk.WorldACL(zk.PermAll), 0)

	fmt.Printf("## [ZN] %+v\n", string(path))
	node, _, _ := z.Get("/test2")
	fmt.Printf("## [ZN] %+v\n", string(node))
}
