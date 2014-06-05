package main

import (
	"fmt"
	"log"
	"sync"
	"time"
	"zk"
)

var (
	TESTIP = []string{
		//"172.19.32.16",
		"192.168.56.101",
	}
)

func main() {
	log.Println("testing")
	conn := zk.New()
	err := conn.Connect(TESTIP)
	if err != nil {
		log.Println(err)
	}
	t1 := time.Now()
	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			traverse(conn, "/ymb/loggers")
		}()
	}
	wg.Wait()
	t2 := time.Now()
	log.Println(t2.Sub(t1))
}

// traverse all znodes under the specified path
func traverse(conn *zk.ZkCli, path string) {
	children, err := conn.Children(path)
	if err != nil {
		return
	}

	if len(children) <= 0 {
		flag, err := conn.Exists(path)
		if err == nil {
			fmt.Println(path, flag)
		}
	}
	for _, znode := range children {
		if path == "/" {
			//fmt.Printf("Searching ZNode: /%s\n", znode)
			traverse(conn, "/"+znode)
		} else {
			//fmt.Printf("Searching ZNode: %s/%s\n", path, znode)
			traverse(conn, path+"/"+znode)
		}
	}
}
