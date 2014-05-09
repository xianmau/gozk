package zk

const (
	// 协议版本
	protocolVersion = 0
	// 默认端口
	DefaultPort = 2181
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

type EventType int32

var (
	eventNames = map[EventType]string{
		EventnodeCreated:         "EventNodeCreated",
		EventnodeDeleted:         "EventNodeDeleted",
		EventnodeDataChanged:     "EventNodeDataChanged",
		EventnodeChildrenChanged: "EventNodeChildrenChanged",
		EventSession:             "EventSession",
		EventNotWatching:         "EventNotWatching",
	}
)

func (t EventType) String() string {
	if name := eventNames[t]; name != "" {
		return name
	}
	return "Unknown"
}

const (
	EventNodeCreate          = EventType(1)
	EventNodeDelete          = EventType(2)
	EventNodeDataChanged     = EventType(3)
	EventNodeChildrenChanged = EventType(4)

	EventSession     = EventType(-1)
	EventNotWatching = EventType(-2)
)
