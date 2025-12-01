package compress

import (
	"crypto/aes"
	"crypto/des"
	"fmt"
)

// 定义支持的密钥长度
const (
	AES128KeyLength = 16 // AES-128
	AES192KeyLength = 24 // AES-192
	AES256KeyLength = 32 // AES-256
	DESKeyLength    = 8  // DES 56位
)

// ValidateKeyLength 校验密钥长度是否符合算法要求
// algorithm: 加密算法
// keyBytes: 密钥字节数组
// return: 合法返回 true, 否则 false
func ValidateKeyLength(algorithm string, keyBytes []byte) bool {
	keyLen := len(keyBytes)
	switch algorithm {
	case "aes":
		return keyLen == AES128KeyLength || keyLen == AES192KeyLength || keyLen == AES256KeyLength
	case "des":
		return keyLen == DESKeyLength
	default:
		return false
	}
}

// PadKey 补全密钥长度, 长度不足时自动补全
// algorithm: 加密算法
// key: 用户输入的密钥字符串
// return: 补全后的密钥字节数组、错误
func PadKey(algorithm string, key string) ([]byte, error) {
	keyBytes := []byte(key)
	switch algorithm {
	case "aes":
		// AES 密钥填充/截断为 32 字节
		if len(keyBytes) == 0 {
			return nil, fmt.Errorf("密钥不能为空")
		}
		if len(keyBytes) < AES256KeyLength {
			padded := make([]byte, AES256KeyLength)
			copy(padded, keyBytes)
			return padded, nil
		}
		return keyBytes[:AES256KeyLength], nil
	case "des":
		// DES 密钥 8 字节
		if len(keyBytes) == 0 {
			return nil, fmt.Errorf("DES 密钥不能为空")
		}
		padded := make([]byte, DESKeyLength)
		copy(padded, keyBytes)
		return padded, nil
	default:
		return nil, fmt.Errorf("不支持的算法: %s", algorithm)
	}
}

// CheckCipherSupport 校验算法是否支持
func CheckCipherSupport(algorithm string) error {
	switch algorithm {
	case "aes":
		// 测试 AES 加密器初始化
		_, err := aes.NewCipher(make([]byte, AES256KeyLength))
		return err
	case "des":
		// 测试 DES 加密器初始化
		_, err := des.NewCipher(make([]byte, DESKeyLength))
		return err
	default:
		return fmt.Errorf("不支持的加密算法: %s, 仅支持 aes/des", algorithm)
	}
}
