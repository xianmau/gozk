package zk

const (
	protocolVersion = 0           // 协议版本
	defaultPort     = 2181        // 默认端口
	bufferSize      = 1024 * 1024 // 缓冲区大小，单位是字节
	sendChanSize    = 16          // 消息队列的大小
)

var (
	// 第一次请求时，客户端发送16位空字节数组到服务器端
	emptyPassword = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
)

// 操作码
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

// 连接状态
const (
	StateUnknown      = -1   // 未知
	StateDisconnected = 0    // 未连接
	StateConnecting   = 1    // 正在连接
	StateConnected    = 100  // 已连接
	StateExpired      = -112 // 超时
	StateHasSession   = 101  // 已有会话
)

// ACL的权限值
const (
	PermRead    = 1 << iota // 读权限，可以获取当前节点的数据，可以列出当前节点所有的子节点
	PermWrite               // 写权限，可以向当前node写数据
	PermCreate              // 创建权限，可以在在当前节点下创建子节点
	PermDeleted             // 删除权限，可以删除当前的节点
	PermAdmin               // 管理权限，可以设置当前节点的权限
	PermAll     = 0x1f      // 所有权限
)
