package main

import (
	"fmt"
	"github.com/samuel/go-zookeeper/zk"
	"time"
	//"net"
)

var (
	TESTIP = []string{
		"172.19.32.39:2181",
	}
)

func main() {
	//fmt.Println("Zookeeper Client Start.")
	//fmt.Println("-----------------------")

	c, _, err := zk.Connect(TESTIP, time.Second)
	if err != nil {
		panic(err)
	}
	//fmt.Printf("## [zk] %+v\n", c)

	children, stat, ch, err := c.ChildrenW("/")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v %+v\n", children, stat)
	e := <-ch
	fmt.Printf("%+v\n", e)
}
