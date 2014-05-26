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

	zk := zk.Connect(TESTIP, time.Second)

	//time.Sleep(time.Second)
	fmt.Printf("## [zk] %+v\n", zk)

	//node, _ := zk.Get("/test")
	//fmt.Printf("## [ZN] %+v\n", node)

	//zk.Create("/test", []byte("test data"))
}
