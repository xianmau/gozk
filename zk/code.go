package zk

import (
	"errors"
)

const (
	opCreate   = 1
	opDelete   = 2
	opExists   = 3
	opGet      = 4
	opSet      = 5
	opChildren = 8
	opPing     = 11
	opClose    = -11
)

const (
	stateDisconnect = 1
	stateConnecting = 2
	stateConnected  = 3
)

const (
	errOk                      = 0
	errAPIError                = -100
	errNoNode                  = -101
	errNoAuth                  = -102
	errBadVersion              = -103
	errNoChildrenForEphemerals = -108
	errNodeExists              = -110
	errNotEmpty                = -111
	errSessionExpired          = -112
	errInvalidCallback         = -113
	errInvalidAcl              = -114
	errAuthFailed              = -115
	errClosing                 = -116
	errNothing                 = -117
	errSessionMoved            = -118
	errConnectionDisabled      = -200 // 连接不可用
	errChannelClosed           = -201 // 请求队列关闭
	errEOF                     = -202 // 读结束
)

var (
	errMap = map[int32]error{
		errOk: nil,
		errConnectionDisabled: errors.New("zk: connection disabled"),
		errChannelClosed:      errors.New("zk: channel closed"),
		errEOF:                errors.New("zk: end of file"),
	}
)
