package main

import (
	"bufio"
	"fmt"
	"github.com/samuel/go-zookeeper/zk"
	"os"
	"strings"
	"time"
)

var (
	TESTIP = []string{
		"172.19.32.16",
		//"192.168.56.101",
	}
)

func main() {

	fmt.Println("You can use the following commands:")
	fmt.Println("exist path")
	fmt.Println("create path data")
	fmt.Println("get path")
	fmt.Println("set path data")
	fmt.Println("ls path")
	fmt.Println("del path")
	fmt.Println("delrec path")
	fmt.Println("Or you will enter 'quit' to exit.")

	conn, _, err := zk.Connect(TESTIP, 1*time.Second)
	if err != nil {
		panic(err)
	}
	//defer conn.Close()

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf(" > ")
		data, _, _ := reader.ReadLine()
		cmd := string(data)
		s := strings.Split(cmd, " ")
		if len(s) == 1 {
			if s[0] == "quit" {
				return
			}
		} else if len(s) == 2 {
			if s[0] == "ls" {
				children, _, err := conn.Children(s[1])
				if err != nil {
					panic(err)
				}
				fmt.Printf("List children of [%d] [%s]: %+v\n", len(children), s[1], children)
			} else if s[0] == "get" {
				data, _, err := conn.Get(s[1])
				if err != nil {
					panic(err)
				}
				fmt.Printf("Data of [%s]: %+v\n", s[1], string(data))
			} else if s[0] == "exist" {
				flag, _, err := conn.Exists(s[1])
				if err != nil {
					panic(err)
				}
				fmt.Printf("[%s] exist: %+v\n", s[1], flag)
			} else if s[0] == "del" {
				err := conn.Delete(s[1], -1)
				if err != nil {
					panic(err)
				}
				fmt.Printf("[%s] delete!\n", s[1])
			}
		} else if len(s) == 3 {
			if s[0] == "create" {
				_, err := conn.Create(s[1], []byte(s[2]), 0, zk.WorldACL(zk.PermAll))
				if err != nil {
					panic(err)
				}
				fmt.Printf("[%s] created!\n", s[1])
			} else if s[0] == "set" {
				_, err := conn.Set(s[1], []byte(s[2]), -1)
				if err != nil {
					panic(err)
				}
				fmt.Printf("[%s] set!\n", s[1])
			}
		}
	}

}
