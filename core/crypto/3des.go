package crypto

import (
	"fmt"
)

// 3DES = DES-EDE3,  3次DES加密, 安全性中等、性能差、实现难度中等
// 密钥长度 24 字节, 112/168 位有效密钥, 兼容 DES, 安全性比 DES 高, 但仍低于 AES, 逐渐被淘汰

// TripleDESCrypter 3DES 加解密器
// 3DES = DES-EDE3，密钥长度 24 字节，兼容 DES，安全性更高
type TripleDESCrypter struct{}

// DoCrypto 3DES 加解密核心逻辑
func (t *TripleDESCrypter) DoCrypto(opts CryptoOptions) error {

	// 1. 校验密钥长度(3DES 要求 24 字节)
	if len(opts.Key) != 24 {
		return fmt.Errorf("3DES 密钥长度必须为 24 字节，当前：%d", len(opts.Key))
	}

	// 2. 初始化 3DES 加密器/解密器
	// 3. 选择模式（CBC/ECB 等）
	// 4. 分块加解密+填充/去填充
	// 5. 写入元数据（IV/盐值）+ 加密数据
	return fmt.Errorf("3DES 算法暂未实现")
}
