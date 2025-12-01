package crypto

import (
	"fmt"
)

// RC4 是一种轻量级流加密, 性能高实现难度低, 但存在大量漏洞, 禁止用于敏感数据

// RC4Crypter RC4 流加密器
type RC4Crypter struct{}

// DoCrypto RC4 加密/解密核心逻辑
func (r *RC4Crypter) DoCrypto(opts CryptoOptions) error {
	// 1. 初始化 RC4 S 盒
	// 2. 流加密/解密
	// 3. 无块大小限制, 直接逐字节处理
	return fmt.Errorf("RC4 算法暂未实现")
}
