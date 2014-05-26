package zk

import (
	"errors"
)

// 错误信息
var (
	ErrUnhandledFieldType = errors.New("zk: unhandled field type")
	ErrPtrExpected        = errors.New("zk: encode/decode expect a non-nil pointer to struct")
	ErrShortBuffer        = errors.New("zk: buffer too small")

	//所有IP不通，说明整个zone挂了
	ErrZoneDown = errors.New("zk: zone down or all ips faild")

	ErrConnectionClosed        = errors.New("zk: connection closed")
	ErrUnknown                 = errors.New("zk: unknown error")
	ErrAPIError                = errors.New("zk: api error")
	ErrNoNode                  = errors.New("zk: node does not exist")
	ErrNoAuth                  = errors.New("zk: not authenticated")
	ErrBadVersion              = errors.New("zk: version conflict")
	ErrNoChildrenForEphemerals = errors.New("zk: ephemeral nodes may not have children")
	ErrNodeExists              = errors.New("zk: node already exists")
	ErrNotEmpty                = errors.New("zk: node has children")
	ErrSessionExpired          = errors.New("zk: session has been expired by the server")
	ErrInvalidCallback         = errors.New("zk: invalid call back")
	ErrInvalidACL              = errors.New("zk: invalid ACL specified")
	ErrAuthFailed              = errors.New("zk: client authentication failed")
	ErrClosing                 = errors.New("zk: zookeeper is closing")
	ErrNothing                 = errors.New("zk: no server responsees to process")
	ErrSessionMoved            = errors.New("zk: session moved to another server, so operation is ignored")

	int32ToError = map[int32]error{
		// 小于0的还没做映射
		errOk:                      nil,
		errAPIError:                ErrAPIError,
		errNoNode:                  ErrNoNode,
		errNoAuth:                  ErrNoAuth,
		errBadVersion:              ErrBadVersion,
		errNoChildrenForEphemerals: ErrNoChildrenForEphemerals,
		errNodeExists:              ErrNodeExists,
		errNotEmpty:                ErrNotEmpty,
		errSessionExpired:          ErrSessionExpired,
		errInvalidCallback:         ErrInvalidCallback,
		errInvalidAcl:              ErrInvalidACL,
		errAuthFailed:              ErrAuthFailed,
		errClosing:                 ErrClosing,
		errNothing:                 ErrNothing,
		errSessionMoved:            ErrSessionMoved,
	}
)

const (
	errOk = 0
	// 系统或服务器的错误
	errSystemError          = -1
	errRuntimeInconsistency = -2
	errDataInconsistency    = -3
	errConnectionLoss       = -4
	errMarshallingError     = -5
	errUnimplemented        = -6
	errOperationTimeout     = -7
	errBadArguments         = -8
	errInvalidState         = -9
	// API错误
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
)
