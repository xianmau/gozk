package main

import (
	"fmt"
	"time"
	"zk"
	//"net"
)

var (
	TESTIP = []string{
		"172.19.32.39:2181",
	}
)

func main() {
	fmt.Println("Zookeeper Client Start.")
	fmt.Println("-----------------------")

	zk, err := zk.Connect(TESTIP, time.Second)
	if err != nil {
		panic(err)
	}
	fmt.Printf("## [zk] %+v\n", zk)

	node := zk.Get("/test")
	fmt.Printf("## [ZN] %+v\n", node)

	zk.Create("/test", []byte("test data"))
}
