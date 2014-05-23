package zk

const (
	protocolVersion = 0           // 协议版本
	defaultPort     = 2181        // 默认端口
	bufferSize      = 1536 * 1024 // 缓冲区大小，单位是字节
)

var (
	// 第一次请求时，客户端发送16位空字节数组到服务器端
	emptyPassword = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
)

const (
	opNotify      = 0
	opCreate      = 1
	opDelete      = 2
	opExists      = 3
	opGetData     = 4
	opSetData     = 5
	opGetAcl      = 6
	opSetAcl      = 7
	opGetChildren = 8
	opSync        = 9

	opPing         = 11
	opGetChildren2 = 12
	opCheck        = 13
	opMulti        = 14

	opClose = -11

	opSetAuth    = 100
	opSetWatches = 101

	opWatchrEvent = -2
)
