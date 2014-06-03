package zk

import (
	"errors"
)

type ErrCode int32

// 错误信息
var (
	ErrUnknown                 = errors.New("zk: unknown error")
	ErrUnhandledFieldType      = errors.New("zk: unhandled field type")
	ErrPtrExpected             = errors.New("zk: encode/decode expect a non-nil pointer to struct")
	ErrShortBuffer             = errors.New("zk: buffer too small")
	ErrConnectionClosed        = errors.New("zk: connection closed")
	ErrDialingFaild            = errors.New("zk: all dialing faild")
	ErrAPIError                = errors.New("zk: api error")
	ErrNoNode                  = errors.New("zk: node does not exist")
	ErrNoAuth                  = errors.New("zk: not authenticated")
	ErrBadVersion              = errors.New("zk: version conflict")
	ErrNoChildrenForEphemerals = errors.New("zk: ephemeral nodes may not have children")
	ErrNodeExists              = errors.New("zk: node already exists")
	ErrNotEmpty                = errors.New("zk: node has children")
	ErrSessionExpired          = errors.New("zk: session has been expired by the server")
	ErrInvalidCallback         = errors.New("zk: callback invalid")
	ErrInvalidACL              = errors.New("zk: invalid ACL specified")
	ErrAuthFailed              = errors.New("zk: client authentication failed")
	ErrClosing                 = errors.New("zk: zookeeper is closing")
	ErrNothing                 = errors.New("zk: no server responsees to process")
	ErrSessionMoved            = errors.New("zk: session moved to another server, so operation is ignored")
)

var (
	errCodeToError = map[ErrCode]error{
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
	errOk                      = 0
	errAPIError                = ErrCode(-100)
	errNoNode                  = ErrCode(-101)
	errNoAuth                  = ErrCode(-102)
	errBadVersion              = ErrCode(-103)
	errNoChildrenForEphemerals = ErrCode(-108)
	errNodeExists              = ErrCode(-110)
	errNotEmpty                = ErrCode(-111)
	errSessionExpired          = ErrCode(-112)
	errInvalidCallback         = ErrCode(-113)
	errInvalidAcl              = ErrCode(-114)
	errAuthFailed              = ErrCode(-115)
	errClosing                 = ErrCode(-116)
	errNothing                 = ErrCode(-117)
	errSessionMoved            = ErrCode(-118)
)
