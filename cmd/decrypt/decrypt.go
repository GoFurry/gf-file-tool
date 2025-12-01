// Package decrypt /cmd/decrypt/decrypt.go
package decrypt

import (
	"github.com/GoFurry/gf-file-tool/cmd"
	"github.com/GoFurry/gf-file-tool/core/crypto"
	"github.com/GoFurry/gf-file-tool/utils"
	"github.com/GoFurry/gf-file-tool/utils/compress"
	"github.com/GoFurry/gf-file-tool/utils/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// decryptCmd 压缩命令实例
var decryptCmd = &cobra.Command{
	Use:   "decrypt [source...]",
	Short: "解密文件/目录",
	Long: `解密文件/目录:
  gf-file-tool decrypt test.enc -k 123456 -o test.txt
  gf-file-tool decrypt ./docs_enc -k 123456 --algorithm aes`,
	Args: cobra.MinimumNArgs(1),
	Run: func(c *cobra.Command, args []string) {
		// 解析参数
		outputPath, _ := c.Flags().GetString("output")
		algorithm, _ := c.Flags().GetString("algorithm")
		key, _ := c.Flags().GetString("key")
		keyLength, _ := c.Flags().GetInt("key-length")
		salt, _ := c.Flags().GetString("salt")

		if salt == "" && algorithm == "aes" {
			log.Error("解密必须指定盐值 (--salt/-s), 请填写加密时生成的盐值")
			return
		}

		// 校验密钥
		if key == "" {
			log.Error("必须指定解密密钥 (--key/-k)")
			return
		}

		// 批量获取文件
		var sourcePaths []string
		for _, src := range args {
			files, err := compress.GetFileList(src)
			if err != nil {
				log.Warn("跳过无效路径:", src, ", 错误:", err)
				continue
			}
			sourcePaths = append(sourcePaths, files...)
		}
		if len(sourcePaths) == 0 {
			log.Error("无有效解密文件")
			return
		}

		// 处理密钥
		if keyLength == 0 {
			if algorithm == "des" {
				keyLength = compress.DESKeyLength
			} else {
				keyLength = compress.AES256KeyLength
			}
		}
		paddedKey, err := compress.PadKey(algorithm, key)
		if err != nil {
			log.Error("密钥处理失败:", err)
			return
		}

		// 逐个解密文件
		for _, src := range sourcePaths {
			dst := outputPath
			if dst == "" {
				if len(src) > 4 && src[len(src)-4:] == ".enc" {
					dst = src[:len(src)-4]
				} else {
					dst = src + ".dec"
				}
			}

			// 构建解密配置
			opts := crypto.CryptoOptions{
				Algorithm:  algorithm,
				Key:        paddedKey,
				KeyLength:  keyLength,
				Salt:       salt,
				IsEncrypt:  false,
				SourcePath: src,
				OutputPath: dst,
			}

			// 执行解密
			if err := crypto.RunCrypto(opts); err != nil {
				log.Error("解密失败:", src, ", 错误:", err)
				continue
			}
			if utils.VerboseMode() {
				log.Success("解密成功:", src, "→", dst)
			}
		}
	},
}

// InitDecrypt 初始化命令
func InitDecrypt() {
	cmd.GetRootCmd().AddCommand(decryptCmd)

	// 注册参数
	decryptCmd.Flags().StringP("output", "o", "", "输出解密文件路径 (批量时为目录)")
	decryptCmd.Flags().StringP("algorithm", "a", "aes", "解密算法 (aes/des/3des/aes-ctr/rc4/chacha20)")
	decryptCmd.Flags().StringP("key", "k", "", "解密密钥 (必填)")
	decryptCmd.Flags().IntP("key-length", "l", 32, "密钥长度 (AES 16/24/32)")
	decryptCmd.Flags().StringP("salt", "s", "", "解密盐值 (必填, 加密时的盐值)")

	// 绑定 Viper
	_ = viper.BindPFlag("decrypt.algorithm", decryptCmd.Flags().Lookup("algorithm"))
	_ = viper.BindPFlag("decrypt.key-length", decryptCmd.Flags().Lookup("key-length"))
	_ = viper.BindPFlag("decrypt.salt", decryptCmd.Flags().Lookup("salt"))
}
