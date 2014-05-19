package zk

type ACL struct {
	Perms  int32
	Scheme string
	Id     string
}

//
func WorldACL(perms int32) []ACL {
	return []ACL{{perms, "world", "anyone"}}
}
