package zk

const (
	PermRead   = 1  // 可获取当前节点的数据及所有子节点
	PermWrite  = 2  // 可向当前节点写数据
	PermCreate = 4  // 可在当前节点下创建子节点
	PermDelete = 8  // 可删除当前的节点
	PermAdmin  = 16 // 可设置当前节点的权限
	PermAll    = 31 // 拥有以上所有权限
)

var (
	WorldACL = []ACL{{PermAll, "world", "anyone"}} // 全局ACL
)

type ACL struct {
	Perms  int32
	Scheme string
	Id     string
}
