package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	"zk"
)

var (
	TESTIP = []string{
		"172.19.32.16",
		//"192.168.56.101",
	}
)

func main() {
	fmt.Println("You can use the following commands:")
	fmt.Println("create path data")
	fmt.Println("set path data")
	fmt.Println("exist path")
	fmt.Println("get path")
	fmt.Println("ls path")
	fmt.Println("del path")
	fmt.Println("Or you will enter 'quit' to exit.")

	conn := zk.Connect(TESTIP, time.Second)
	defer conn.Close()

	go func() {
		err := conn.DeleteRecur("/ymb")
		if err != nil {
			panic(err)
		}
		if flag, err := conn.Exists("/ymb"); err == nil && !flag {
			conn.Create("/ymb", "", zk.WorldACL(zk.PermAll), 0)
		}
		for i := 1; i < 1500; i++ {
			if flag, err := conn.Exists("/ymb/" + strconv.Itoa(i)); err == nil && !flag {
				conn.Create("/ymb/"+strconv.Itoa(i), "", zk.WorldACL(zk.PermAll), 0)
			}
			fmt.Println(i)
		}
	}()

	reader := bufio.NewReader(os.Stdin)
	for {
		data, _, _ := reader.ReadLine()
		cmd := string(data)
		s := strings.Split(cmd, " ")
		if len(s) == 1 {
			if s[0] == "quit" {
				return
			}
		} else if len(s) == 2 {
			if s[0] == "ls" {
				children, err := conn.Children(s[1])
				if err != nil {
					panic(err)
				}
				fmt.Printf("List children of [%s]: %+v\n", s[1], children)
			} else if s[0] == "get" {
				data, err := conn.Get(s[1])
				if err != nil {
					panic(err)
				}
				fmt.Printf("Data of [%s]: %+v\n", s[1], string(data))
			} else if s[0] == "exist" {
				flag, err := conn.Exists(s[1])
				if err != nil {
					panic(err)
				}
				fmt.Printf("[%s] exist: %+v\n", s[1], flag)
			} else if s[0] == "del" {
				err := conn.Delete(s[1])
				if err != nil {
					panic(err)
				}
				fmt.Printf("[%s] delete!\n", s[1])
			}
		} else if len(s) == 3 {
			if s[0] == "create" {
				err := conn.Create(s[1], s[2], zk.WorldACL(zk.PermAll), 0)
				if err != nil {
					panic(err)
				}
				fmt.Printf("[%s] created!\n", s[1])
			} else if s[0] == "set" {
				err := conn.Set(s[1], s[2])
				if err != nil {
					panic(err)
				}
				fmt.Printf("[%s] set!\n", s[1])
			}
		}
	}

}
