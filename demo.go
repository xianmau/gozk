package main

import (
	"encoding/binary"
	"fmt"
	"time"
)

func main() {
	fmt.Printf("%d", time.Duration(1))
	b := []byte{0, 0, 0, 4, 5}
	fmt.Println(BytesToInt64(b))
}
func BytesToInt32(buf []byte) int32 {
	return int32(binary.BigEndian.Uint32(buf))
}

func BytesToInt64(buf []byte) int64 {
	return int64(binary.BigEndian.Uint64(buf))
}

func BytesToBool(buf []byte) bool {
	if buf[0] == 1 {
		return true
	}
	return false
}
