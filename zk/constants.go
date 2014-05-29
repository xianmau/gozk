package zk

const (
	protocolVersion = 0           // 协议版本
	defaultPort     = 2181        // 默认端口
	bufferSize      = 1024 * 1024 // 缓冲区大小，单位是字节，这里是1M
	sendChanSize    = 16          // 信道的缓冲大小
)

var (
	// 第一次请求时，客户端发送16位空字节数组到服务器端
	emptyPassword = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
)
