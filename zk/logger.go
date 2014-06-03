package zk

import (
	"log"
	"os"
)

var logger *log.Logger

func init() {
	logger = log.New(os.Stderr, "[log]", log.Ldate|log.Ltime|log.Lshortfile)
}
