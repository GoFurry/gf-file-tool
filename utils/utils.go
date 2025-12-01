// Package utils /utils/utils.go
package utils

import (
	"crypto/rand"
	"math/big"

	"github.com/spf13/viper"
)

// 通用工具

// QuietMode 校验静默模式开启状态
func QuietMode() bool {
	val := viper.Get("quiet")
	if b, ok := val.(bool); ok {
		return b
	}
	return false
}

// VerboseMode 校验详细日志模式
func VerboseMode() bool {
	val := viper.Get("verbose")
	if b, ok := val.(bool); ok {
		return b
	}
	return false
}

// GenerateRandomString 生成指定长度的随机字符串
func GenerateRandomString(length int) (string, error) {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			return "", err
		}
		result[i] = chars[num.Int64()]
	}
	return string(result), nil
}
