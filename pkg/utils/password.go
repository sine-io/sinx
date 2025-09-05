package utils

import (
	"crypto/md5"
	"fmt"
)

// HashPassword 对密码进行哈希处理 (简单实现，生产环境建议使用bcrypt)
func HashPassword(password string) (string, error) {
	hash := md5.Sum([]byte(password))
	return fmt.Sprintf("%x", hash), nil
}

// CheckPassword 验证密码
func CheckPassword(password, hash string) bool {
	hashedPassword, err := HashPassword(password)
	if err != nil {
		return false
	}
	return hashedPassword == hash
}
