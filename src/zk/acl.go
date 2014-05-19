package zk

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
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
