package zk

type State int32

// 连接状态
const (
	StateUnknown      = -1   // 未知
	StateDisconnected = 0    // 未连接
	StateConnecting   = 1    // 正在连接
	StateConnected    = 100  // 已连接
	StateExpired      = -112 // 超时
	StateHasSession   = 101  // 已有会话
)
