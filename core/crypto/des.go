package crypto

import (
	"crypto/cipher"
	"crypto/des"
	"crypto/rand"
	"fmt"
	"io"
	"os"

	"github.com/GoFurry/gf-file-tool/progress"
	"github.com/GoFurry/gf-file-tool/utils"
	"github.com/GoFurry/gf-file-tool/utils/compress"
	"github.com/GoFurry/gf-file-tool/utils/log"
)

// DES 安全性极低, 性能中等, 实现难度较低
// 56 位有效密钥, 可被轻易暴力破解, 已经被淘汰, 仅适合教学演示

// DESCrypter DES 加解密器
type DESCrypter struct{}

// DoCrypto DES 加密/解密核心逻辑
func (d *DESCrypter) DoCrypto(opts CryptoOptions) error {
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

	// 初始化 DES 加密器/解密器
	block, err := des.NewCipher(opts.Key)
	if err != nil {
		return fmt.Errorf("初始化 DES 失败: %v (DES 仅支持 8 字节密钥) ", err)
	}

	// 进度条
	bar := progress.NewFileProgressBar(fileInfo.Size(), opts.SourcePath)
	defer progress.FinishProgress(bar)

	if opts.IsEncrypt {
		// DES 加密
		iv := make([]byte, des.BlockSize)
		if _, err := io.ReadFull(rand.Reader, iv); err != nil {
			return fmt.Errorf("生成 IV 失败: %v", err)
		}

		// 先写入 IV + 盐值
		if _, err := dstFile.Write(iv); err != nil {
			return fmt.Errorf("写入 IV 失败: %v", err)
		}
		saltBytes, _ := compress.ParseSalt(opts.Salt)
		if _, err := dstFile.Write(saltBytes); err != nil {
			return fmt.Errorf("写入盐值失败: %v", err)
		}

		// 读取所有原始数据
		var plainData []byte
		buf := make([]byte, 4*1024*1024) // 4MB 缓冲区
		totalRead := int64(0)
		for {
			n, err := srcFile.Read(buf)
			if err != nil && err != io.EOF {
				return fmt.Errorf("读取源文件失败: %v", err)
			}
			if n == 0 {
				break
			}
			plainData = append(plainData, buf[:n]...)
			totalRead += int64(n)
			if bar != nil {
				_ = bar.Set64(totalRead)
			}
		}

		// 整体 PKCS7 填充
		padder := NewPKCS7Padding(des.BlockSize)
		paddedData := padder.Pad(plainData)

		// CBC 模式分块加密写入
		mode := cipher.NewCBCEncrypter(block, iv)
		encryptBuf := make([]byte, len(paddedData))
		mode.CryptBlocks(encryptBuf, paddedData)

		// 写入加密后的数据
		if _, err := dstFile.Write(encryptBuf); err != nil {
			return fmt.Errorf("写入加密数据失败: %v", err)
		}

	} else {
		// DES 解密
		// 读取 IV + 盐值
		iv := make([]byte, des.BlockSize)
		if _, err := io.ReadFull(srcFile, iv); err != nil {
			return fmt.Errorf("读取 IV 失败: %v (文件可能不是 DES 加密或已损坏) ", err)
		}
		saltBytes := make([]byte, compress.DefaultSaltLength)
		if _, err := io.ReadFull(srcFile, saltBytes); err != nil {
			return fmt.Errorf("读取盐值失败: %v", err)
		}

		// 读取所有加密数据
		var cipherData []byte
		buf := make([]byte, 4*1024*1024)
		totalRead := int64(len(iv) + len(saltBytes))
		for {
			n, err := srcFile.Read(buf)
			if err != nil && err != io.EOF {
				return fmt.Errorf("读取加密文件失败: %v", err)
			}
			if n == 0 {
				break
			}
			cipherData = append(cipherData, buf[:n]...)
			totalRead += int64(n)
			if bar != nil {
				_ = bar.Set64(totalRead)
			}
		}

		// 校验加密数据长度
		if len(cipherData)%des.BlockSize != 0 {
			return fmt.Errorf("加密数据长度非法: %d 字节 (必须是 8 字节倍数), 文件可能损坏", len(cipherData))
		}

		// CBC 整体解密
		mode := cipher.NewCBCDecrypter(block, iv)
		plainData := make([]byte, len(cipherData))
		mode.CryptBlocks(plainData, cipherData)

		// 整体 PKCS7 去填充
		padder := NewPKCS7Padding(des.BlockSize)
		unpaddedData, err := padder.Unpad(plainData)
		if err != nil {
			return fmt.Errorf("去填充失败: %v (密钥错误、盐值错误或文件损坏)", err)
		}

		// 写入解密后的数据
		if _, err := dstFile.Write(unpaddedData); err != nil {
			return fmt.Errorf("写入解密数据失败: %v", err)
		}
	}

	// 强制刷盘
	if err := dstFile.Sync(); err != nil && utils.VerboseMode() {
		log.Warn("刷盘失败: %v", err)
	}

	return nil
}

// PKCS7Padding PKCS7 填充器
type PKCS7Padding struct {
	blockSize int
}

// NewPKCS7Padding 初始化 PKCS7 填充器
func NewPKCS7Padding(blockSize int) *PKCS7Padding {
	return &PKCS7Padding{blockSize: blockSize}
}

// Pad 整体填充
func (p *PKCS7Padding) Pad(data []byte) []byte {
	if len(data) == 0 {
		// 空数据填充为一个块
		padding := p.blockSize
		padText := make([]byte, padding)
		for i := range padText {
			padText[i] = byte(padding)
		}
		return padText
	}
	padding := p.blockSize - len(data)%p.blockSize
	padText := make([]byte, padding)
	for i := range padText {
		padText[i] = byte(padding)
	}
	return append(data, padText...)
}

// Unpad 整体去填充
func (p *PKCS7Padding) Unpad(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, fmt.Errorf("空数据无法去填充")
	}
	if length%p.blockSize != 0 {
		return nil, fmt.Errorf("数据长度非法: %d 字节 (必须是 %d 字节倍)", length, p.blockSize)
	}
	padding := int(data[length-1])
	if padding <= 0 || padding > p.blockSize {
		return nil, fmt.Errorf("无效的填充长度: %d (必须 1-%d)", padding, p.blockSize)
	}
	// 校验所有填充字节
	for i := 0; i < padding; i++ {
		if data[length-1-i] != byte(padding) {
			return nil, fmt.Errorf("填充字节不合法: 位置 %d 应为 %d, 实际 %d", length-1-i, padding, data[length-1-i])
		}
	}
	return data[:length-padding], nil
}
