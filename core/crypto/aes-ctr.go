package crypto

import (
	"fmt"
)

// AES-CTR 安全性/性能与 AES 一致, 流模式无填充泄露风险, 比 CBC 方式更安全

// AESCTRCrypter AES-CTR 模式加解密器
type AESCTRCrypter struct{}

// DoCrypto AES-CTR 加密/解密核心逻辑
func (a *AESCTRCrypter) DoCrypto(opts CryptoOptions) error {
	// 1. 初始化 AES-CTR 模式
	// 2. 流加密/解密
	// 3. 写入 Counter/盐值 等元数据
	// 4. 分块流式处理
	return fmt.Errorf("AES-CTR 算法暂未实现")
}
