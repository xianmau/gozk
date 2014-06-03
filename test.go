package main

import (
	"log"
	"zk"
)

var (
	// 测试用的IP
	TESTIP = []string{"192.168.56.101:2181"}
)

func main() {
	zkCli := zk.New()

	err := zkCli.Connect(TESTIP)
	handleErr(err)

	err = zkCli.ping()
	handleErr(err)

	flag, err := zkCli.exists("/ymb")
	log.Println(flag, err)

	err = zkCli.close()

	err = zkCli.connect(server)
	data, err := zkCli.get("/ymb")
	log.Println(string(data), err)

	children, err := zkCli.children("/")
	log.Println(children, err)

	err = zkCli.create("/test2", []byte("goodgood"))
	log.Println(err)

	err = zkCli.set("/test", []byte("nonono"))
	log.Println(err)

	err = zkCli.delete("/test/okok")
	err = zkCli.delete("/test")
	log.Println(err)
}

func handleErr(err error) {
	if err != nil {
		log.Printf("%v\n", err)
		panic(err)
	}
}
