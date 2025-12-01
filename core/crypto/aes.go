package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"os"

	"github.com/GoFurry/gf-file-tool/progress"
	"github.com/GoFurry/gf-file-tool/utils/compress"
)

// AES 安全性高、性能高、实现难度中等, 是目前主流的加密方式
// 128 位密钥已足够安全, 256 位适合高安全场景, 是当前行业标准

// AESCrypter AES 加解密器
type AESCrypter struct{}

// DoCrypto AES 加密/解密核心逻辑
func (a *AESCrypter) DoCrypto(opts CryptoOptions) error {
	// 打开源文件
	srcFile, err := os.Open(opts.SourcePath)
	if err != nil {
		return fmt.Errorf("打开源文件失败: %v", err)
	}
	defer srcFile.Close()

	// 创建输出文件
	dstFile, err := os.Create(opts.OutputPath)
	if err != nil {
		return fmt.Errorf("创建输出文件失败: %v", err)
	}
	defer dstFile.Close()

	// 获取文件大小
	fileInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("获取文件信息失败: %v", err)
	}

	// 初始化 AES 加密器/解密器
	block, err := aes.NewCipher(opts.Key)
	if err != nil {
		return fmt.Errorf("初始化 AES 失败: %v", err)
	}

	// GCM 模式
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("初始化 GCM 模式失败: %v", err)
	}

	// 进度条
	bar := progress.NewFileProgressBar(fileInfo.Size(), opts.SourcePath)
	defer progress.FinishProgress(bar)

	if opts.IsEncrypt {
		// AES 加密流程
		// 生成随机 nonce
		nonce := make([]byte, gcm.NonceSize())
		if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
			return fmt.Errorf("生成 nonce 失败: %v", err)
		}

		// 先写入 nonce + 盐值
		if _, err := dstFile.Write(nonce); err != nil {
			return fmt.Errorf("写入 nonce 失败: %v", err)
		}
		saltBytes, _ := compress.ParseSalt(opts.Salt)
		if _, err := dstFile.Write(saltBytes); err != nil {
			return fmt.Errorf("写入盐值失败: %v", err)
		}

		// 分块加密写入
		buf := make([]byte, 4*1024*1024) // 4MB 缓冲区
		totalWritten := int64(0)
		for {
			n, err := srcFile.Read(buf)
			if err != nil && err != io.EOF {
				return fmt.Errorf("读取文件失败: %v", err)
			}
			if n == 0 {
				break
			}

			// 加密
			cipherText := gcm.Seal(nil, nonce, buf[:n], nil)
			// 写入
			if _, err := dstFile.Write(cipherText); err != nil {
				return fmt.Errorf("写入加密数据失败: %v", err)
			}

			// 更新进度
			totalWritten += int64(n)
			if bar != nil {
				_ = bar.Set64(totalWritten)
			}
		}
	} else {
		// AES 解密流程
		// 先读取 nonce + 盐值
		nonce := make([]byte, gcm.NonceSize())
		if _, err := io.ReadFull(srcFile, nonce); err != nil {
			return fmt.Errorf("读取 nonce 失败: %v (文件可能不是 AES 加密或已损坏)", err)
		}
		saltBytes := make([]byte, compress.DefaultSaltLength)
		if _, err := io.ReadFull(srcFile, saltBytes); err != nil {
			return fmt.Errorf("读取盐值失败: %v", err)
		}

		// 分块解密写入
		buf := make([]byte, 4*1024*1024+gcm.Overhead()) // 预留 GCM 额外长度
		totalRead := int64(len(nonce) + len(saltBytes))
		for {
			n, err := srcFile.Read(buf)
			if err != nil && err != io.EOF {
				return fmt.Errorf("读取加密文件失败: %v", err)
			}
			if n == 0 {
				break
			}

			// 解密
			plainText, err := gcm.Open(nil, nonce, buf[:n], nil)
			if err != nil {
				return fmt.Errorf("解密失败: %v (密钥/盐值错误或文件损坏)", err)
			}
			// 写入
			if _, err := dstFile.Write(plainText); err != nil {
				return fmt.Errorf("写入解密数据失败: %v", err)
			}

			// 更新进度
			totalRead += int64(n)
			if bar != nil {
				_ = bar.Set64(totalRead)
			}
		}
	}

	return nil
}
