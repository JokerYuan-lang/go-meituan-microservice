package utils

import (
	"math/rand"
	"time"
)

// RandomString 生成指定长度的随机数字字符串
func RandomString(length int) string {
	rand.Seed(time.Now().UnixNano())
	digits := "0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = digits[rand.Intn(len(digits))]
	}
	return string(b)
}

// ContainsString 判断字符串是否在切片中
func ContainsString(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
