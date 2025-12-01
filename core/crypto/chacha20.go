package crypto

import (
	"fmt"
)

// ChaCha20是谷歌设计的一种现代流加密, 安全性与 AES-256 相当, 无已知有效破解手段, 实现难度较高

// ChaCha20Crypter ChaCha20 流加密器
type ChaCha20Crypter struct{}

// DoCrypto ChaCha20 加密/解密核心逻辑
func (c *ChaCha20Crypter) DoCrypto(opts CryptoOptions) error {
	// 1. 初始化 ChaCha20 密钥/nonce
	// 2. 流加密/解密
	// 3. 支持 256 位密钥, 安全性高
	return fmt.Errorf("ChaCha20 算法暂未实现")
}
