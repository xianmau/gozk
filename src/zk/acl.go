package zk

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
)

// ACL的权限值
const (
	PermsCreate  = iota //创建权限，可以在在当前节点下创建子节点
	PermsDeleted = iota //删除权限，可以删除当前的节点
	PermsRead    = iota //读权限，可以获取当前节点的数据，可以列出当前节点所有的子节点
	PermsWrite   = iota //写权限，可以向当前node写数据
	PermsAdmin   = iota //管理权限，可以设置当前节点的权限
)

// 访问控制的结构
type ACL struct {
	Perms  int32
	Scheme string
	Id     string
}

// 代表某一特定的用户（客户端）
func WorldACL(perms int32) []ACL {
	return []ACL{{perms, "world", "anyone"}}
}

// 代表任何已经通过验证的用户（客户端）
func AuthACL(perms int32) []ACL {
	return []ACL{{perms, "auth", ""}}
}

// 通过用户名和密码进行认证
func DigestACL(perms int32, user, password string) []ACL {
	userPass := []byte(fmt.Sprintf("%s:%s", user, password))
	h := sha1.New()
	if n, err := h.Write(userPass); err != nil || n != len(userPass) {
		panic("SHA1 failed.")
	}
	digest := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return []ACL{{perms, "digest", fmt.Sprintf("%s:%s", user, digest)}}
}

// 通过客户端IP地址进行认证
func IpACL(perms int32, ip string) []ACL {
	return []ACL{{perms, "ip", ip}}
}

// 超级管理员认证
func SuperACL(perms int32, super string) []ACL {
	return []ACL{{perms, "super", super}}
}
