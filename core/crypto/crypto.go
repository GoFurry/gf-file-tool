package crypto

import (
	"crypto/sha256"
	"fmt"

	"github.com/GoFurry/gf-file-tool/utils/compress"
	"golang.org/x/crypto/pbkdf2"
)

//入门级：DES、RC4（逻辑简单，无复杂模式）；
//进阶级：AES（CBC/GCM 模式）、3DES（多轮加密）；
//专家级：AES-CTR（计数器管理）、ChaCha20（矩阵运算、nonce 管理）；

// CryptoOptions 加解密配置
type CryptoOptions struct {
	Algorithm  string // 算法（aes/des）
	Key        []byte // 原始密钥
	KeyLength  int    // 密钥长度（AES：16/24/32；DES：8）
	Salt       string // 盐值（空则自动生成）
	IsEncrypt  bool   // 加密/解密
	SourcePath string // 源文件路径
	OutputPath string // 输出文件路径
}

// Crypter 加解密接口
type Crypter interface {
	DoCrypto(opts CryptoOptions) error
}

// NewCrypter 扩展支持更多算法
func NewCrypter(algorithm string) (Crypter, error) {
	switch algorithm {
	case "aes":
		return &AESCrypter{}, nil
	case "des":
		return &DESCrypter{}, nil
	case "3des": // 3DES（Triple DES）
		return &TripleDESCrypter{}, nil
	case "aes-ctr": // AES-CTR 模式
		return &AESCTRCrypter{}, nil
	case "rc4": // RC4 流加密
		return &RC4Crypter{}, nil
	case "chacha20": // ChaCha20 流加密
		return &ChaCha20Crypter{}, nil
	default:
		return nil, fmt.Errorf("不支持的算法：%s（仅支持 aes/des/3des/aes-ctr/rc4/chacha20）", algorithm)
	}
}

// DeriveKey 密钥派生（适配 DES 8 字节密钥）
func DeriveKey(rawKey []byte, salt []byte, keyLength int) []byte {
	// 强制 DES 密钥长度为 8 字节
	if keyLength != 8 && len(rawKey) > 0 {
		// 先判断算法？不，直接根据 keyLength 强制修正
		if keyLength == 0 || keyLength == 32 { // 若传入 AES 默认值，且是 DES 算法
			// 这里通过 rawKey 长度辅助判断（DES 原始密钥应处理为 8 字节）
			if len(rawKey) == compress.DESKeyLength {
				keyLength = compress.DESKeyLength
			}
		}
	}
	// PBKDF2 配置：迭代次数 10000，哈希算法 SHA256
	return pbkdf2.Key(rawKey, salt, 10000, keyLength, sha256.New)
}

// RunCrypto 统一加解密入口（完善盐值/密钥逻辑）
func RunCrypto(opts CryptoOptions) error {
	// 1. 参数校验
	if opts.Algorithm == "" {
		opts.Algorithm = "aes" // 默认 AES
	}
	if len(opts.Key) == 0 {
		return fmt.Errorf("密钥不能为空")
	}
	if !compress.CheckPathExist(opts.SourcePath) {
		return fmt.Errorf("源文件不存在：%s", opts.SourcePath)
	}

	// 2. 盐值处理
	var saltBytes []byte
	if opts.IsEncrypt {
		// 加密：自动生成盐值
		if opts.Salt == "" {
			salt, err := compress.GenerateSalt(compress.DefaultSaltLength)
			if err != nil {
				return fmt.Errorf("生成盐值失败：%v", err)
			}
			opts.Salt = salt
			fmt.Printf("[Info] 自动生成盐值：%s\n", opts.Salt) // 打印盐值，解密时需要
		}
		saltBytes, _ = compress.ParseSalt(opts.Salt)
	} else {
		// 解密：盐值由用户传入（加密时打印的盐值）
		if opts.Salt == "" {
			return fmt.Errorf("解密必须指定盐值（--salt/-s），加密时生成的盐值：%s", opts.Salt)
		}
		saltBytes, _ = compress.ParseSalt(opts.Salt)
	}

	// 3. 密钥派生（适配 DES 8 字节）
	keyLength := opts.KeyLength
	if keyLength == 0 {
		if opts.Algorithm == "des" {
			keyLength = 8 // DES 固定 8 字节
		} else {
			keyLength = compress.AES256KeyLength // AES 默认 32 字节
		}
	}
	derivedKey := DeriveKey(opts.Key, saltBytes, keyLength)

	// 4. 校验密钥长度
	if !compress.ValidateKeyLength(opts.Algorithm, derivedKey) {
		return fmt.Errorf("密钥长度不合法：%d 字节（算法：%s，要求：AES(16/24/32)、DES(8)）", len(derivedKey), opts.Algorithm)
	}

	// 5. 创建加解密器
	crypter, err := NewCrypter(opts.Algorithm)
	if err != nil {
		return err
	}

	// 6. 执行加解密
	opts.Key = derivedKey
	return crypter.DoCrypto(opts)
}
