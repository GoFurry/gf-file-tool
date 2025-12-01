package compress

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// DefaultSaltLength 默认盐值长度
const DefaultSaltLength = 16

// GenerateSalt 生成随机盐值
// length: 盐值字节长度
// return: 十六进制编码的盐值字符串、错误
func GenerateSalt(length int) (string, error) {
	if length <= 0 {
		length = DefaultSaltLength
	}

	// 生成随机字节
	saltBytes := make([]byte, length)
	_, err := rand.Read(saltBytes)
	if err != nil {
		return "", fmt.Errorf("生成盐值失败: %v", err)
	}

	// 转为十六进制字符串
	return hex.EncodeToString(saltBytes), nil
}

// ParseSalt 将十六进制盐值字符串转回字节数组
func ParseSalt(saltStr string) ([]byte, error) {
	saltBytes, err := hex.DecodeString(saltStr)
	if err != nil {
		return nil, fmt.Errorf("解析盐值失败: %v", err)
	}
	return saltBytes, nil
}
